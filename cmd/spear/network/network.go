package network

import (
	"bytes"
	"errors"
	"fmt"
	"net"
    "log"

	"../crypto"

	"github.com/emirpasic/gods/sets/treeset"
)

const minimumPacketBufferSize = 5

// Addr is a container of candidates and the current one
type Addr struct {
	current    *net.UDPAddr
	Candidates []*net.UDPAddr
}

//Bind tries to bind the client to one of the address
func (addr *Addr) Bind() (*net.UDPConn, error) {
	for _, cand := range addr.Candidates {
		conn, err := net.ListenUDP("udp", cand)
		if err == nil {
			addr.current = cand
            log.Println("Successfully bound to", cand.String())
			return conn, nil
		} else {
            log.Println("Attempted to bind to", cand.String())
        }
	}
	return nil, errors.New("Unable to bind to any candidates")
}

//Set sets the current UDPAddr used by the Client
func (addr *Addr) Set(currentNew *net.UDPAddr) error {
	for _, cand := range addr.Candidates {
		if cand.IP.Equal(currentNew.IP) && cand.Port == currentNew.Port {
			addr.current = currentNew
			return nil
		}
	}
	return errors.New("UDPAddr is not in candidate list")
}

//Get returns the current address
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

//Packet refers to a decrypted incoming packet sent by a peer
type Packet struct {
	ID   []byte
	Data []byte
}

//Peer refers to another spear user
type Peer struct {
	PublicKey []byte
	Addr      Addr

	receivedPackets *treeset.Set
}

//Client refers to the backend of the client containing all basic information needed by the core
type Client struct {
	SecretKey []byte
	nonce     []byte

	PeerList []*Peer

	Addr Addr
	conn *net.UDPConn
}

//Initialize setup the client, should be called first
func (client *Client) Initialize() error {
	crypto.Init()
	conn, err := client.Addr.Bind()
	if err != nil {
		return err
	}
	conn.SetReadBuffer(0x100000)
	client.conn = conn
	client.nonce = crypto.RandomBytes(crypto.NonceSize)
	return nil
}

//Start starts listening to incoming packets
func (client *Client) Start() {
	for {
		buffer := make([]byte, 0x1000)
		size, addr, err := client.conn.ReadFromUDP(buffer)
		if err != nil {
            log.Panicln(err)
			continue
		}

		pk, id, plaintext, err := crypto.DecryptBytes(buffer[:size], client.SecretKey)
		if err != nil {
            log.Panicln(err)
			continue
		}

		sender := client.GetPeerByPublicKey(pk)
		if err := sender.Addr.Set(addr); err != nil {
			panic(err)
		} else {
			if sender == nil {
				sender = &Peer{
					PublicKey: pk,
					Addr: Addr{
						Candidates: []*net.UDPAddr{addr},
					},
				}
				sender.Addr.Set(addr)
			}
		}
		if sender.receivedPackets == nil {
			sender.receivedPackets = treeset.NewWith(sortByPacketID)
		}
		sender.receivedPackets.Add(Packet{
			ID:   id,
			Data: plaintext,
		})
	}
}

//SendBytes send a packet to another peer, bytes should be unencrypted.
func (client *Client) SendBytes(peer *Peer, bytes []byte) {
	ciphertext := crypto.EncryptBytes(peer.PublicKey, client.SecretKey, bytes, client.nonce)
	peer.Addr.Write(client.conn, ciphertext)
}

//GetPeerByPublicKey attempts to find a peer with the corresponding public key. Returns nil if nothing is found.
func (client *Client) GetPeerByPublicKey(pk []byte) *Peer {
	for _, peer := range client.PeerList {
		if bytes.Compare(pk, peer.PublicKey) == 0 {
			return peer
		}
	}
	return nil
}

//GetNewPacket returns a new packet in the peer buffer
func (peer *Peer) GetNewPacket() *Packet {
	if peer.receivedPackets == nil {
		return nil
	}
	if peer.receivedPackets.Size() < minimumPacketBufferSize {
		return nil
	}
	packet := peer.receivedPackets.Values()[0].(Packet)
	peer.receivedPackets.Remove(packet)
	return &packet
}

func sortByPacketID(a, b interface{}) int {
	c1 := a.(Packet)
	c2 := b.(Packet)

	return bytes.Compare(c1.ID, c2.ID)
}
