// +build linux darwin

package main

import (
	"log"
	"os/exec"
)

func System(c string) {
	cmd := exec.Command("/bin/sh", "-c", c)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("tun device up failed with %s\n", err)
	}
}
