package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"

	"github.com/songgao/water"
)

var clientAddr *net.UDPAddr
var tcpListener *net.TCPListener

func localToRemoteS(conn interface{}) {
	packet := make([]byte, 1500-20-8)
	for {
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
					continue
				}
			}
		case *net.TCPConn:
			{

			}
		}

	}
}

func remoteToLocalS(conn interface{}) {
	packet := make([]byte, 1500-20-8)
	for {
		var n int
		var err error
		var peerAddr *net.UDPAddr
		switch conn.(type) {
		case *net.UDPConn:
			{
				n, peerAddr, err = udpConn.ReadFromUDP(packet)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		case *net.TCPConn:
			{
				n, peerAddr, err = udpConn.ReadFromUDP(packet)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}

		if n < 2+12+16 {
			continue
		}

		err = decrypt(packet[12+2:n], []byte(EncryptionKey), packet[2:2+12])

		if err != nil {
			continue
		}

		clientAddr = peerAddr

		if err != nil {
			continue
		}

		_, err = tun.Write(packet[2+12 : n-16])
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func acceptFromRemore() {
	for {
		tcpConn, err := tcpListener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go remoteToLocalS(tcpConn)
		go localToRemoteS(tcpConn)
	}
}

func RunServer() {
	var err error
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "dsvpn"
	tun, err = water.New(config)

	if err != nil {
		log.Fatal(err)
	}

	System(fmt.Sprintf("ip link set dev %s up", config.Name))
	System(fmt.Sprintf("ip addr add %s peer %s dev %s", ServerTunIP, ClientTunIP, config.Name))
	System(fmt.Sprintf("ifconfig %s mtu %d", config.Name, 1500-20-8-2-12-16))

	log.Printf("TUN Interface UP, Name: %s\n", tun.Name())

	udpConn, err = net.ListenUDP("udp", ServerAddr.(*net.UDPAddr))
	if err != nil {
		log.Fatal(err)
	}
	go remoteToLocalS(udpConn)
	go localToRemoteS(udpConn)

	tcpListener, err = net.ListenTCP("tcp", ServerAddr.(*net.TCPAddr))
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
