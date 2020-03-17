package crypto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
)

//NonceSize is the size of nonce used by EncryptBytes
const NonceSize = chacha20poly1305.NonceSize

//EncryptBytes creates an encrypted packet and conumes the plaintext storage
func EncryptBytes(otherPk, userSk, plaintext []byte, packetID uint32) []byte {
	id := uint32ToByte(packetID)

	ckey, mkey := createTimeBaseKey(otherPk, userSk, 0)
	nonce := mac512(mkey, id)[:NonceSize]

	cipher, err := chacha20poly1305.New(ckey)
	if err != nil {
		panic(err)
	}

	ciphertext := cipher.Seal([]byte{}, nonce, plaintext, []byte{})
	return append(id, ciphertext...)
}

//DecryptBytes takes an encrypted packet and reutrns (packet id, plaintext)
func DecryptBytes(c, otherPk, userSk []byte) (uint32, []byte, error) {
	reader := bytes.NewReader(c)
	id := make([]byte, 4)
	reader.Read(id)
	packet := make([]byte, reader.Len())
	reader.Read(packet)

	for _, offset := range []int64{0, -1, 1} {
		ckey, mkey := createTimeBaseKey(otherPk, userSk, offset)
		cipher, err := chacha20poly1305.New(ckey)
		if err != nil {
			panic(err)
		}
		nonce := mac512(mkey, id)[:NonceSize]
		if plaintext, err := cipher.Open([]byte{}, nonce, packet, []byte{}); err == nil {
			return byteToUint32(id), plaintext, nil
		}
	}
	return 0, nil, errors.New("Unable to decrypt messsage")
}

func createTimeBaseKey(otherPk, userSk []byte, offset int64) ([]byte, []byte) {
	seed := createKeySeed(otherPk, userSk)
	value := make([]byte, 8)
	binary.LittleEndian.PutUint64(value, uint64(time.Now().UTC().Unix()/30+offset))
	key := mac512(seed, value)
	return key[0:32], key[32:64]
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
