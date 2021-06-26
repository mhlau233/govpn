package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/songgao/water"
)

var tun *water.Interface

func localToRemoteC(conn interface{}, ctx context.Context, cancel context.CancelFunc, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	// udp payload must <= 1500-20-8
	packet := make([]byte, 1500-20-8)
	for {
		if ctx.Err() != nil {
			return
		}
		// read from local tun
		n, err := tun.Read(packet[2+12:])
		if err != nil {
			log.Fatal(err)
		}

		// wrap into protocol
		rand.Read(packet[2 : 2+12])

		// encrypt data from the tun only
		encrypt(packet[2+12:2+12+n], []byte(EncryptionKey), packet[2:2+12])

		switch c := conn.(type) {
		case *net.UDPConn:
			{
				// 16 bytes tag will be added at the tail after encryption
				_, err = c.Write(packet[0 : 2+12+n+16])
				if err != nil {
					log.Println(err)
					continue
				}
			}
		case *net.TCPConn:
			{
				binary.BigEndian.PutUint16(packet[0:2], uint16(12+n+16))
				_, err = c.Write(packet[0 : 2+12+n+16])
				if err != nil {
					cancel()
					return
				}
			}
		}

	}
}

func remoteToLocalC(conn interface{}, ctx context.Context, cancel context.CancelFunc, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	packet := make([]byte, 1500-20-8)
	packetHeader := make([]byte, 2)
	for {
		if ctx.Err() != nil {
			return
		}
		var n int
		var err error
		switch c := conn.(type) {
		case *net.UDPConn:
			{
				n, err = c.Read(packet)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		case *net.TCPConn:
			{
				// read header first
				_, err := io.ReadFull(c, packetHeader)
				if err != nil {
					log.Println(err)
					cancel()
					return
				}
				len := binary.BigEndian.Uint16(packetHeader)
				_, err = io.ReadFull(c, packet[2:2+len])
				if err != nil {
					log.Println(err)
					cancel()
					return
				}
				n = int(len + 2)
			}
		}

		// tun payload must be at least 1 byte, so any udp packet < 12 + 16 bytes is abnormal
		if n < 2+12+16 {
			continue
		}

		decrypt(packet[2+12:n], []byte(EncryptionKey), packet[2:2+12])

		// 16 bytes tag must be trimmed before injecting into tun
		_, err = tun.Write(packet[2+12 : n-16])
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func RunClient() {
	var err error
	tun, err = InitTun()

	if err != nil {
		log.Fatal(err)
	}

	switch runtime.GOOS {
	case "darwin":
		System(fmt.Sprintf("ifconfig %s %s %s up", tun.Name(), ClientTunIP, ServerTunIP))
		// tun mtu must be 1500-20-8-(12+16)
		// data read from tun must <= 1500 - IPLen - UDPHeaderlen - EncapsulationLen
		System(fmt.Sprintf("ifconfig %s mtu %d", tun.Name(), 1500-20-8-2-12-16))
	case "linux":
		System(fmt.Sprintf("ip link set dev %s up", tun.Name()))
		System(fmt.Sprintf("ip addr add %s peer %s dev %s", ClientTunIP, ServerTunIP, tun.Name()))
		System(fmt.Sprintf("ifconfig %s mtu %d", tun.Name(), 1500-20-8-2-12-16))
	case "windows":
		System(fmt.Sprintf("netsh interface ip set address name=\"%s\" static %s 255.255.255.0 %s metric=automatic", tun.Name(), ClientTunIP, ServerTunIP))
		System(fmt.Sprintf("netsh interface ipv4 set subinterface \"%s\" mtu=%d", tun.Name(), 1500-20-8-2-12-16))
	}

	log.Printf("TUN Interface UP, Name: %s\n", tun.Name())

	if ProtocolType == "udp" {
		go runUDP()
	} else {
		go runTCP()
	}
	select {}
}

func runUDP() {
	udpConn, err := net.DialUDP("udp", nil, ServerAddr.(*net.UDPAddr))
	if err != nil {
		log.Fatalf("failed to dialup udp, %s", err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	ctx, cancel := context.WithCancel(context.Background())
	go remoteToLocalC(udpConn, ctx, cancel, &wg)
	go localToRemoteC(udpConn, ctx, cancel, &wg)
	wg.Wait()
}

func runTCP() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", ServerAddr.String())
	for {
		var tcpConn *net.TCPConn
		var err error
		for {
			tcpConn, err = net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				log.Println("failed to dialup tcp")
				time.Sleep(5 * time.Second)
				continue
			}
			break
		}
		var wg sync.WaitGroup
		wg.Add(2)
		ctx, cancel := context.WithCancel(context.Background())
		go remoteToLocalC(tcpConn, ctx, cancel, &wg)
		go localToRemoteC(tcpConn, ctx, cancel, &wg)
		wg.Wait()
		log.Println("conn failed, reconnecting")
	}
}
