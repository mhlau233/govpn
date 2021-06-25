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
