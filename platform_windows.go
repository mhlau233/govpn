// +build windows

package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/songgao/water"
)

func System(c string) {
	cmd := exec.Command("cmd", "/c", c)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("tun device up failed with %s\n", err)
	}
}

func InitTun() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.PlatformSpecificParams = water.PlatformSpecificParams{
		ComponentID: "TAP0901",
		Network:     fmt.Sprintf("%s/24", ClientTunIP),
	}
	return water.New(config)
}
