package crypto

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/curve25519"
)

//CreatePublicKey generates a public key give a secret key sk
func CreatePublicKey(sk []byte) []byte {
	pk, err := curve25519.X25519(sk, curve25519.Basepoint)
	if err != nil {
		panic("Unable to create public key")
	}
	return pk
}

//RandomBytes return a []byte of n size with random content
func RandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic("RandomBytes error!")
	}
	return b
}

//BytesIncr treats a byte array as a big-endian number and increase it by one
func BytesIncr(b []byte) []byte {
	z := new(big.Int)
	z.SetBytes(b)
	z.Add(z, big.NewInt(1))
	return z.Bytes()
}

func uint32ToByte(i uint32) []byte {
	body := make([]byte, 4)
	binary.LittleEndian.PutUint32(body, i)
	return body
}

func hash512(message []byte) []byte {
	hash, err := blake2b.New512(nil)
	if err != nil {
		panic("Cannot create blake2b.New512")
	}
	hash.Write(message)
	return hash.Sum([]byte{})
}

func mac512(key []byte, message []byte) []byte {
	hash, err := blake2b.New512(key)
	if err != nil {
		panic("Cannot create blake2b.New512")
	}
	hash.Write(message)
	return hash.Sum([]byte{})
}
