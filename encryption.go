package main

import (
	"crypto/aes"
	"crypto/cipher"
)

// encrypt plaintext inplace and add 16 bytes padding on tail
func encrypt(plaintext, key, nonce []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	aesgcm.Seal(plaintext[:0], nonce, plaintext, nil)
	return nil
}

// decrypt ciphertext inplace, ciphertext must contains 16 bytes tag
func decrypt(ciphertext, key, nonce []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	_, err = aesgcm.Open(ciphertext[:0], nonce, ciphertext, nil)
	if err != nil {
		return err
	}
	return nil
}
