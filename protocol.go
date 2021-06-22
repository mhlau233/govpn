package main

type Protocol struct {
	Length  uint16
	Nonce   [12]byte
	Payload []byte
}
