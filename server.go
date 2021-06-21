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

func localToRemoteS() {
	packet := make([]byte, 1500-20-8)
	for {
		n, err := tun.Read(packet[12:])
		if err != nil {
			log.Fatal(err)
			os.Exit(-1)
		}

		rand.Read(packet[0:12])

		encrypt(packet[12:12+n], []byte(EncryptionKey), packet[0:12])

		_, err = conn.WriteToUDP(packet[0:n+12+16], clientAddr)
		if err != nil {
			continue
		}
	}
}

func remoteToLocalS() {
	packet := make([]byte, 1500-20-8)
	for {
		n, peerAddr, err := conn.ReadFromUDP(packet)
		if err != nil {
			log.Println(err)
			continue
		}

		if n < 12+16 {
			continue
		}

		err = decrypt(packet[12:n], []byte(EncryptionKey), packet[0:12])

		if err != nil {
			continue
		}

		clientAddr = peerAddr

		if err != nil {
			continue
		}

		_, err = tun.Write(packet[12 : n-16])
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
	config.Name = "dsvpn"
	tun, err = water.New(config)

	if err != nil {
		log.Fatal(err)
	}

	System(fmt.Sprintf("ip link set dev %s up", config.Name))
	System(fmt.Sprintf("ip addr add %s peer %s dev %s", ServerTunIP, ClientTunIP, config.Name))
	System(fmt.Sprintf("ifconfig %s mtu %d", config.Name, 1500-20-8-12-16))

	log.Printf("TUN Interface UP, Name: %s\n", tun.Name())

	conn, err = net.ListenUDP("udp", ServerAddr)
	if err != nil {
		log.Fatal(err)
	}

	go remoteToLocalS()
	go localToRemoteS()
	select {}
}
