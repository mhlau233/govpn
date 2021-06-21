package main

type Protocol struct {
	Nonce   [12]byte
	Payload []byte
}
