package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/songgao/water"
)

var tun *water.Interface
var conn *net.UDPConn

func localToRemoteC() {

	// udp payload must <= 1500-20-8
	packet := make([]byte, 1500-20-8)
	for {
		// read from local tun
		n, err := tun.Read(packet[12:])
		if err != nil {
			log.Fatal(err)
		}

		// wrap into protocol
		rand.Read(packet[0:12])

		// encrypt data from the tun only
		encrypt(packet[12:12+n], []byte(EncryptionKey), packet[0:12])

		// 16 bytes tag will be added at the tail after encryption
		n, err = conn.Write(packet[0 : n+12+16])
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func remoteToLocalC() {
	packet := make([]byte, 1500-20-8)
	for {
		n, err := conn.Read(packet)
		if err != nil {
			log.Println(err)
			continue
		}
		// tun payload must be at least 1 byte, so any udp packet < 12 + 16 bytes is abnormal
		if n < 12+16 {
			continue
		}

		decrypt(packet[12:n], []byte(EncryptionKey), packet[0:12])

		// 16 bytes tag must be trimmed before injecting into tun
		n, err = tun.Write(packet[12 : n-16])
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func RunClient() {
	var err error
	tun, err = water.New(water.Config{
		DeviceType: water.TUN,
	})

	if err != nil {
		log.Fatal(err)
	}

	System(fmt.Sprintf("ifconfig %s %s %s up", tun.Name(), ClientTunIP, ServerTunIP))
	// tun mtu must be 1500-20-8-(12+16)
	// data read from tun must <= 1500 - IPLen - UDPHeaderlen - EncapsulationLen
	System(fmt.Sprintf("ifconfig %s mtu %d", tun.Name(), 1500-20-8-12-16))

	log.Printf("TUN Interface UP, Name: %s\n", tun.Name())

	conn, err = net.DialUDP("udp", nil, ServerAddr)
	if err != nil {
		log.Fatal("failed to dialup udp")
	}

	go remoteToLocalC()
	go localToRemoteC()
	select {}
}
