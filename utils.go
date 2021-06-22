package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
)

func System(c string) {
	cmd := exec.Command("/bin/sh", "-c", c)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("tun device up failed with %s\n", err)
	}
}

func parseFlags() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s: \n", os.Args[0])
		flag.PrintDefaults()
	}
	var ServerEndpoint string
	flag.StringVar(&Mode, "m", "", "run mode")
	flag.StringVar(&ServerEndpoint, "se", "", "server endpoint")
	flag.StringVar(&ClientTunIP, "cip", "10.0.0.100", "client tun ip")
	flag.StringVar(&ServerTunIP, "sip", "10.0.0.1", "server tun ip")
	flag.StringVar(&EncryptionKey, "k", "abcdefgqywuwyw", "encryption key")

	flag.Parse()

	if Mode == "" {
		fmt.Println("-m is required [client, server]")
		os.Exit(-1)
	}

	if ServerTunIP == "" {
		fmt.Println("-sip is required")
		os.Exit(-1)
	}

	if ClientTunIP == "" {
		fmt.Println("-sip is required")
		os.Exit(-1)
	}

	if ServerEndpoint == "" {
		fmt.Println("-se is required")
		os.Exit(-1)
	}

	var err error
	ServerAddr, err = net.ResolveUDPAddr("udp", ServerEndpoint)
	if err != nil {
		fmt.Println("failed to parse se")
		os.Exit(-1)
	}

	fmt.Printf("sip: %s cip: %s\n", ServerTunIP, ClientTunIP)
	fmt.Printf("server endpoint: %s\n", ServerAddr.String())
}
