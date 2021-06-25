// +build linux

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"

	"github.com/songgao/water"
)

var clientAddr *net.UDPAddr

var mutex sync.Mutex
var CurrentContext context.Context
var currentContextCancel context.CancelFunc

func localToRemoteS(conn interface{}, ctx context.Context, cancel context.CancelFunc, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	defer waitGroup.Done()

	packet := make([]byte, 1500-20-8)
	for {
		if ctx.Err() != nil {
			return
		}
		n, err := tun.Read(packet[2+12:])
		if err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}

		rand.Read(packet[2 : 2+12])

		encrypt(packet[2+12:2+12+n], []byte(EncryptionKey), packet[2:2+12])

		switch c := conn.(type) {
		case *net.UDPConn:
			{
				_, err = c.WriteToUDP(packet[0:2+12+n+16], clientAddr)
				if err != nil {
					cancel()
					return
				}
			}
		case *net.TCPConn:
			{
				binary.BigEndian.PutUint16(packet[0:2], uint16(12+n+16))
				_, err := c.Write(packet[0 : 2+12+n+16])
				if err != nil {
					c.Close()
					cancel()
					return
				}
			}
		}
	}
}

func remoteToLocalS(conn interface{}, ctx context.Context, cancel context.CancelFunc, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	defer waitGroup.Done()

	packet := make([]byte, 1500-20-8)
	packetHeader := make([]byte, 2)
	for {
		var n int
		var err error
		var peerAddr *net.UDPAddr
		switch c := conn.(type) {
		case *net.UDPConn:
			{
				if ctx.Err() != nil {
					return
				}
				n, peerAddr, err = c.ReadFromUDP(packet)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		case *net.TCPConn:
			{
				if ctx.Err() != nil {
					c.Close()
					return
				}
				// read header first
				_, err := io.ReadFull(c, packetHeader)
				if err != nil {
					log.Println(err)
					c.Close()
					cancel()
					return
				}
				len := binary.BigEndian.Uint16(packetHeader)
				if len > 1500-20-8-2 {
					log.Printf("len:%d > 1500-20-8-2", len)
					c.Close()
					cancel()
					return
				}
				_, err = io.ReadFull(c, packet[2:2+len])
				if err != nil {
					log.Println(err)
					c.Close()
					cancel()
					return
				}
				n = int(len + 2)
			}
		}

		if n < 2+12+16 {
			continue
		}

		err = decrypt(packet[12+2:n], []byte(EncryptionKey), packet[2:2+12])

		if err != nil {
			log.Println("decrypt error")
			continue
		}

		mutex.Lock()
		if CurrentContext != ctx {
			if currentContextCancel != nil {
				currentContextCancel()
			}
			CurrentContext = ctx
			currentContextCancel = cancel
			// log.Printf("localToRemoteS with %v %v", waitGroup, conn)
			go localToRemoteS(conn, ctx, cancel, waitGroup)
		}
		mutex.Unlock()

		clientAddr = peerAddr

		_, err = tun.Write(packet[2+12 : n-16])
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func RunServer() {
	var err error
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "govpn"
	tun, err = water.New(config)

	if err != nil {
		log.Fatal(err)
	}

	System(fmt.Sprintf("ip link set dev %s up", config.Name))
	System(fmt.Sprintf("ip addr add %s peer %s dev %s", ServerTunIP, ClientTunIP, config.Name))
	System(fmt.Sprintf("ifconfig %s mtu %d", config.Name, 1500-20-8-2-12-16))

	log.Printf("TUN Interface UP, Name: %s\n", tun.Name())

	go acceptFromRemoteUDP()
	go acceptFromRemoteTCP()

	select {}
}

func acceptFromRemoteTCP() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", ServerAddr.String())
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		tcpConn, err := tcpListener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go func() {
			var wg sync.WaitGroup
			ctx, cancel := context.WithCancel(context.Background())
			go remoteToLocalS(tcpConn, ctx, cancel, &wg)
			wg.Wait()
		}()
	}
}

func acceptFromRemoteUDP() {
	udpConn, err := net.ListenUDP("udp", ServerAddr.(*net.UDPAddr))
	if err != nil {
		log.Fatal(err)
	}
	for {
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())
		go remoteToLocalS(udpConn, ctx, cancel, &wg)
		wg.Wait()
	}
}
