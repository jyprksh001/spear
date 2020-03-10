package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
)

const (
	KEY_BYTES_N int = 32
)

// []byte but with a custom json marshal function for decoding/encoding base64
type Key struct {
	Bytes []byte
}

func (key *Key) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(key.Bytes))
}

func (key *Key) UnmarshalJSON(inBytes []byte) error {
	var encoded string
	if err := json.Unmarshal(inBytes, &encoded); err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}

	if len(decoded) != KEY_BYTES_N {
		return fmt.Errorf("base64: '%s', decoded byte size is %d != %d", encoded, len(decoded), KEY_BYTES_N)
	}

	key.Bytes = decoded
	return nil
}

type Peer struct {
	PK   Key `json:"pk"`
	Host string
	// A list of possible ports
	Ports []int
	//Port the peer is currently using
	CurrentAddr *net.UDPAddr
	IsNotKnown  bool
}
