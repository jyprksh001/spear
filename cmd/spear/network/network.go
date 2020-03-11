package network

import (
	"bytes"
	"errors"
	"log"
	"net"
	"time"

	"../crypto"
)

const minimumPacketBufferSize = 5
const maximumPacketBufferSize = 10
const maximumTimeDifference = 1000

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
		}
		log.Println("Attempted to bind to", cand.String())
	}
	return nil, errors.New("Unable to bind to any candidates")
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
	if len(client.Addr.Candidates) == 0 {
		panic("Address candidates is empty")
	}
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
func (client *Client) Start(stop *bool, done chan bool) {
	for !*stop {
		buffer := make([]byte, 0x1000)
		size, addr, err := client.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println(err)
			continue
		}

		pk, id, plaintext, err := crypto.DecryptBytes(buffer[:size], client.SecretKey)
		if err != nil {
			log.Println(err)
			continue
		}

		sender := client.GetPeerByPublicKey(pk)
		if sender == nil {
			sender = &Peer{
				PublicKey: pk,
				Addr: Addr{
					current: addr,
				},
			}
		}
		if sender.receivedPackets == nil {
			sender.receivedPackets = []*Packet{}
		}
		sender.receivePacket(&Packet{
			ID:           id,
			RawData:      plaintext,
			ReceivedTime: time.Now().UTC().UnixNano() / 1000000,
		})
	}
	done <- true
}

//SendBytes send a packet to another peer, bytes should be unencrypted.
func (client *Client) SendBytes(peer *Peer, bytes []byte) {
	ciphertext := crypto.EncryptBytes(peer.PublicKey, client.SecretKey, bytes, client.nonce)
	peer.Addr.Write(client.conn, ciphertext)
	client.nonce = crypto.BytesIncr(client.nonce)
}

//SendBytesToAll send a packet to all peers, bytes should be unencrypted.
func (client *Client) SendBytesToAll(bytes []byte) {
	for _, peer := range client.PeerList {
		client.SendBytes(peer, bytes)
	}
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
