// +build linux darwin

package main

import (
	"log"
	"os/exec"

	"github.com/songgao/water"
)

func System(c string) {
	cmd := exec.Command("/bin/sh", "-c", c)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("tun device up failed with %s\n", err)
	}
}

func InitTun() (*water.Interface, error) {
	return water.New(water.Config{
		DeviceType: water.TUN,
	})
}
