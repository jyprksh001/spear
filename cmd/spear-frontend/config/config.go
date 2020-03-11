package config

import (
    "encoding/json"
    "encoding/base64"
    "strings"
    "strconv"
    "fmt"
    "net"
    "io/ioutil"
)

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


type Key struct {
    Bytes []byte
}

func (key *Key) MarshalJSON() ([]byte, error) {
    return json.Marshal(base64.StdEncoding.EncodeToString(key.Bytes))
}

func (key *Key) UnmarshalJSON(bytes []byte) error {
    fmt.Println("What")

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

type Address struct {
    UDP net.UDPAddr
}

func (addr *Address) MarshalJSON() ([]byte, error) {
    return json.Marshal(addr.UDP.String())
}

func (addr *Address) UnmarshalJSON(bytes []byte) error {
    fmt.Println("WHAT")
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

type Peer struct {
    PublicKey Key `json:"pk"`
    Candidates []Address `json:"candidates"`
}

type Config struct {
    SecretKey Key `json:"sk"`
    Candidates []Address `json:"candidates"`
    Peers []Peer `json:"peers"`
}

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
