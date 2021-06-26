// +build windows

package main

import (
	"fmt"
	"log"
	"os/exec"
)

func System(c string) {
	cmd := exec.Command("cmd", "/c", c)
	str, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(str))
		log.Fatalf("tun device up failed with %s\n", err)
	}
}

func InitTun() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.PlatformSpecificParams = water.PlatformSpecificParams{
		ComponentID: "TAP0901",
		Network:     "10.0.0.100/32",
	}
	return water.New(config)
}
