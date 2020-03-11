package network

import (
    "errors"
    "net"
    "github.com/google/go-cmp/cmp"
    "bytes"
    "fmt"
    "../crypto"
)

const (
    DEFAULT_PORT int = 57975
)

// Addr: a container of candidates and the current one
type Addr struct {
    current *net.UDPAddr
    Candidates []*net.UDPAddr
}

func (addr *Addr) Bind() (*net.UDPConn, error) {
    for _, cand := range addr.Candidates {
        conn, err := net.ListenUDP("udp", cand)
        if err == nil {
            addr.current = cand
            return conn, nil
        }
    }
    return nil, errors.New("Unable to bind to any candidates")
}

func (addr *Addr) Set(currentNew *net.UDPAddr) error {
    for _, cand := range addr.Candidates {
        if cmp.Equal(cand, currentNew) {
            addr.current = currentNew
            return nil
        }
    }
    return errors.New("UDPAddr is not in candidate list")
}

func (addr *Addr) Get() *net.UDPAddr {
    return addr.current
}

func (addr *Addr) Write(conn *net.UDPConn, data []byte) {
    if addr.current == nil {
        for _, cand := range addr.Candidates {
            conn.WriteToUDP(data, cand)
        }
    } else {
        conn.WriteToUDP(data, addr.current)
    }
}

type Packet struct {
    ID uint64
    Data []byte
}

type Peer struct {
    PublicKey []byte
    Addr Addr
}

type Client struct {
    SecretKey []byte
    Nonce []byte

    PeerList []*Peer

    Addr Addr
    conn *net.UDPConn

    Callback *func(*Peer, *Packet)
}

func (client *Client) Initialize() error {
    conn, err := client.Addr.Bind()
    if err != nil {
        return err
    }
    conn.SetReadBuffer(0x100000)
    client.conn = conn
    return nil
}

func (client *Client) Start() {
    for {
        buffer := make([]byte, 0x1000)
        size, addr, err := client.conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Println(err)
            continue
        }

        pk, id, plaintext, err := crypto.DecryptBytes(buffer[:size], client.SecretKey)
        if err != nil {
            fmt.Println(err)
            continue
        }

        sender, ok := client.GetPeerByPublicKey(pk)
        if err := sender.Addr.Set(addr); err != nil {
            panic(err)
        } else {
            if !ok {
                sender = &Peer{
                    PublicKey: pk,
                    Addr: Addr{
                        Candidates: []*net.UDPAddr{addr},
                    },
                }
                sender.Addr.Set(addr)
            }
        }

        go (*client.Callback)(sender, &Packet{
            ID: id,
            Data: plaintext,
        })
    }
}

func (client *Client) SendBytes(peer *Peer, bytes []byte) {
    ciphertext := crypto.EncryptBytes(peer.PublicKey, client.SecretKey, bytes, client.Nonce)
    peer.Addr.Write(client.conn, ciphertext)
}

func (client *Client) GetPeerByPublicKey(pk []byte) (*Peer, bool) {
    for _, peer := range client.PeerList {
        if bytes.Compare(pk, peer.PublicKey) == 0 {
            return peer, true
        }
    }
    return nil, false
}
