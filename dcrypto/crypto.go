// Package dcrypto ...
package dcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"strings"
)

const (
	defaultPrefix = "dcrypto-"

	padingSize = aes.BlockSize - sha1.Size%aes.BlockSize
)

var (
	errInvalidPrefix     = errors.New("invalid prefix, it should be " + defaultPrefix)
	errInvalidCipherText = errors.New("too short cipher string")

	padingBytes = bytes.Repeat([]byte{byte(padingSize)}, padingSize)
)

// Encrypt return enc str
func Encrypt(plainText, key string) (string, error) {
	keyBytes := hash(key)
	plainBytes := []byte(plainText)
	c, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	blockSize := c.BlockSize()
	cipherBytes := make([]byte, blockSize+len(plainBytes))
	_, err = io.ReadFull(rand.Reader, cipherBytes[:blockSize])
	if err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(c, cipherBytes[:blockSize])
	stream.XORKeyStream(cipherBytes[blockSize:], plainBytes)
	return defaultPrefix + hex.EncodeToString(cipherBytes), nil
}

// Decrypt return ori str
func Decrypt(cipherText, key string) (string, error) {
	keyBytes := hash(key)
	if len(cipherText) < 8 || !strings.HasPrefix(cipherText, defaultPrefix) {
		return "", errInvalidPrefix
	}
	cipherBytes := []byte(cipherText)
	cipherBytes = cipherBytes[8:]
	cipherBytes, err := hex.DecodeString(string(cipherBytes))
	if err != nil {
		return "", err
	}
	c, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	blockSize := c.BlockSize()
	if len(cipherBytes) < blockSize {
		return "", errInvalidCipherText
	}
	plainBytes := make([]byte, len(cipherBytes)-blockSize)
	stream := cipher.NewCFBDecrypter(c, cipherBytes[:blockSize])
	stream.XORKeyStream(plainBytes, cipherBytes[blockSize:])
	return string(plainBytes), nil
}

const n = aes.BlockSize - sha1.Size%aes.BlockSize

func hash(src string) []byte {
	sh := sha1.New()
	sh.Write([]byte(src))
	return append(sh.Sum(nil), padingBytes...)
}
