package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
<<<<<<< HEAD

/*
example 'config.json':
{
    "pk": "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=",
    "sk": "ABECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=",
    "ports": [15124, 15125, 15126],
    "peers": [
        {
            "pk": "ACECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=",
            "host": "192.168.1.4",
            "ports": [2313, 2314, 2315]
        },
        {
            "pk": "ADECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=",
            "host": "foo.bar.baz.qux.quux.quuux",
            "ports": [1234, 2345]
        }
    ]
}
*/

type Config struct {
	// the user's keys
	PK Key `json:"pk"`
	SK Key `json:"sk"`
	// A list of ports to bind to
	Ports []int
	// the peer list
	Peers []Peer
}

func (conf *Config) ReadFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	json.Unmarshal(data, conf)
	return nil
}
=======
>>>>>>> 5388b46e06eab348111520d5e8b3c26321b8b416
