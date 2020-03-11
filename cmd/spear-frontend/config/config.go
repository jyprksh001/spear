package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
    "../../spear/network"
)

//ParseAddr turns a string in ipv4:port format to a UDPAddr
func ParseAddr(str string) (*net.UDPAddr, error) {
	result := strings.Split(str, ":")

	var ipstr, portstr string
	if len(result) == 2 {
		ipstr = result[0]
		portstr = result[1]
	} else {
		return nil, fmt.Errorf("Failed to parse Address: %#v, length of split result is %d", str, len(result))
	}

	ip := net.ParseIP(ipstr)
	if ip == nil {
		return nil, fmt.Errorf("Failed to parse IP: %#v", ipstr)
	}

	port, err := strconv.Atoi(portstr)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse Port: %#v", portstr)
	}

	return &net.UDPAddr{IP: ip, Port: port}, nil
}

//Key refers to a public key or a secret key, used for marshalling
type Key struct {
	Bytes []byte
}

//MarshalJSON is used for JSON marshalling
func (key *Key) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(key.Bytes))
}

//UnmarshalJSON is used for JSON unmarshalling
func (key *Key) UnmarshalJSON(bytes []byte) error {

	var encoded string
	if err := json.Unmarshal(bytes, &encoded); err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	if len(decoded) != 32 {
		return fmt.Errorf("Decoded key has byte length %d", len(decoded))
	}

	key.Bytes = decoded
	return nil
}

//Address is a proxy struct used for JSON marshalling
type Address struct {
	UDP net.UDPAddr
}

//MarshalJSON is used for JSON marshalling
func (addr *Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.UDP.String())
}

//UnmarshalJSON is used for JSON unmarshalling
func (addr *Address) UnmarshalJSON(bytes []byte) error {
	var str string
	if err := json.Unmarshal(bytes, &str); err != nil {
		return err
	}

	a, err := ParseAddr(str)
	if err != nil {
		return err
	}
	addr.UDP.IP = a.IP
	addr.UDP.Port = a.Port
	return nil
}

//Peer is a struct used for JSON marshalling, not to be confused with network.Peer
type Peer struct {
	PublicKey  Key       `json:"pk"`
	Candidates []Address `json:"candidates"`
}

func (peer *Peer) LoadToPeer(dest *network.Peer) {
    dest.PublicKey = peer.PublicKey.Bytes
    for _, cand := range peer.Candidates {
        dest.Addr.Candidates = append(dest.Addr.Candidates, &cand.UDP)
    }
}

//Config is a struct used for JSON marshalling, mirroring the structure of config.json
type Config struct {
	SecretKey  Key       `json:"sk"`
	Candidates []Address `json:"candidates"`
	Peers      []Peer    `json:"peers"`
}

//ReadFile reads a config.json and init the values of conf
func (conf *Config) ReadFile(path string) error {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &conf); err != nil {
		return err
	}
	return nil
}

func (conf *Config) LoadToClient(client *network.Client) {
    client.SecretKey = conf.SecretKey.Bytes
    for _, cand := range conf.Candidates {
        client.Addr.Candidates = append(client.Addr.Candidates, &cand.UDP)
    }
    for _, peer := range conf.Peers {
        p := new(network.Peer)
        peer.LoadToPeer(p)
        client.PeerList = append(client.PeerList, p)
    }
}
