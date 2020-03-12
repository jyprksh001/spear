package crypto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"

	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
)

//NonceSize is the size of nonce used by EncryptBytes
const NonceSize = chacha20poly1305.NonceSize

//EncryptBytes creates an encrypted packet and conumes the plaintext storage
func EncryptBytes(otherPk, userSk, plaintext []byte, packetID uint32) []byte {
	id := uint32ToByte(packetID)

	//Length info
	length := uint32(len(plaintext) & 0xFFFF)
	metadata := make([]byte, 2)
	binary.LittleEndian.PutUint32(metadata, length)
	plaintext = append(metadata, plaintext...)

	ciphertext, mkey := encrypt(otherPk, userSk, plaintext, id, 0)

	ciphertext = append(id, ciphertext...)

	//Create MAC
	mac := mac512(mkey, ciphertext)[:2]
	return append(mac, ciphertext...)
}

//DecryptBytes takes an encrypted packet and reutrns (packet id, plaintext)
func DecryptBytes(c, otherPk, userSk []byte) (uint32, []byte, error) {
	reader := bytes.NewReader(c)
	mac := make([]byte, 2)
	id := make([]byte, 4)
	packet := make([]byte, len(c))

	reader.Read(mac)
	reader.Read(id)
	reader.Read(packet)

	for _, offset := range []int64{0, -1, 1} {
		plaintext, mkey := encrypt(otherPk, userSk, packet, id, offset)
		length := binary.LittleEndian.Uint32(plaintext[:2])
		plaintext = plaintext[2 : length+2]

		//Verify MAC
		if bytes.Compare(mac, mac512(mkey, append(id, plaintext...))[:2]) == 0 {
			return binary.LittleEndian.Uint32(id), plaintext, nil
		}
	}
	return 0, nil, errors.New("Unable to decrypt messsage")
}

//Does encryption / decryption and returns the result text and a mac key
func encrypt(otherPk, userSk, text, id []byte, offset int64) ([]byte, []byte) {
	//Generate 512 time based key
	ckey, mkey := createTimeBaseKey(otherPk, userSk, offset)

	//Nonce generator
	nonce := mac512(mkey, id)[:chacha20.NonceSize]

	cipher, err := chacha20.NewUnauthenticatedCipher(ckey, nonce)
	if err != nil {
		panic(err)
	}

	//Unauthenticated encryption
	ciphertext := make([]byte, len(text))
	cipher.XORKeyStream(ciphertext, text)
	return ciphertext, mkey
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
