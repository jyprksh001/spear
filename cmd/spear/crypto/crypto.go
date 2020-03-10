package crypto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/big"
	"math/rand"
	"time"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
)

//NonceSize is the size of nonce used by EncryptBytes
const NonceSize = chacha20poly1305.NonceSize

//EncryptBytes creates an encrypted packet and conumes the plaintext storage
func EncryptBytes(otherPk, userSk, plaintext, nonce []byte) []byte {
	if len(nonce) != chacha20poly1305.NonceSize {
		panic("Nonce size not equals to chacha20poly1305.NonceSize")
	}
	key := createTimeBaseKey(otherPk, userSk, 0)
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		panic("Error setting up chacha20poly1305 cipher")
	}
	userPk := CreatePublicKey(userSk)

	ciphertext := aead.Seal(plaintext[:0], nonce, plaintext, userPk)

	return append(userPk, append(nonce, ciphertext...)...)
}

//DecryptBytes takes an encrypted packet and reutrns (sender pk, packet id, plaintext)
func DecryptBytes(c, userSk []byte) ([]byte, uint64, []byte, error) {
	reader := bytes.NewReader(c)
	pk := make([]byte, 32)
	nonce := make([]byte, chacha20poly1305.NonceSize)
	packet := make([]byte, len(c))

	reader.Read(pk)
	reader.Read(nonce)
	n, _ := reader.Read(packet)
	packet = packet[:n]
	for _, offset := range []int64{0, -1, 1} {
		key := createTimeBaseKey(pk, userSk, offset)
		aead, err := chacha20poly1305.New(key)
		if err != nil {
			panic("Error setting up chacha20poly1305 cipher")
		}

		plaintext, err := aead.Open([]byte{}, nonce, packet, pk)
		if err == nil {
			id := binary.BigEndian.Uint64(nonce[chacha20poly1305.NonceSize-8:])
			return pk, id, plaintext, nil
		}
	}
	return nil, 0, nil, errors.New("Unable to decrypt")
}

func createTimeBaseKey(otherPk, userSk []byte, offset int64) []byte {
	seed := createKeySeed(otherPk, userSk)
	value := make([]byte, 8)
	binary.LittleEndian.PutUint64(value, uint64(time.Now().UTC().Unix()/30+offset))
	return mac256(seed, value)
}

func createKeySeed(otherPk, userSk []byte) []byte {
	userPk := CreatePublicKey(userSk)
	secret, err := curve25519.X25519(userSk, otherPk)
	if err != nil {
		panic("Key exchanged failed")
	}

	var pkconcat []byte
	if bytes.Compare(otherPk, userPk) >= 0 {
		pkconcat = append(otherPk, userPk...)
	} else {
		pkconcat = append(userPk, otherPk...)
	}

	secret = hash512(append(secret, pkconcat...))
	return secret
}

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

//Init sets up the crypto library
func Init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func hash512(message []byte) []byte {
	hash, err := blake2b.New512(nil)
	if err != nil {
		panic("Cannot create blake2b.New512")
	}
	hash.Write(message)
	return hash.Sum([]byte{})
}

func mac256(key []byte, message []byte) []byte {
	hash, err := blake2b.New256(key)
	if err != nil {
		panic("Cannot create blake2b.New256")
	}
	hash.Write(message)
	return hash.Sum([]byte{})
}
