package main

import (
	"log"
	"net"
)

// for client
var Mode string
var ServerTunIP string
var ClientTunIP string
var EncryptionKey string

var ServerAddr *net.UDPAddr

func main() {
	parseFlags()
	switch Mode {
	case "client":
		RunClient()
	case "server":
		RunServer()
	default:
		log.Println("unknown mode")
	}
}
