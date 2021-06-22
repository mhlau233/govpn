package main

import (
	"log"
	"net"
)

var Mode string
var ServerTunIP string
var ClientTunIP string
var EncryptionKey string

var ServerAddr net.Addr

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
