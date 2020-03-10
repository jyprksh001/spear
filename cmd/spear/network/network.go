package network

import (
	"bytes"
	"errors"
	"fmt"
	"net"

	"../crypto"
)

//Client represents a spear client with basic information
type Client struct {
	SecretKey []byte
	Ports     []uint16
	PeerList  []Peer

	//Callback handles incoming packets (packet sender, packet id, content)
	Callback *func(*Peer, uint64, []byte)

	conn  *net.UDPConn
	nonce []byte
}

//Peer refers to another spear user
type Peer struct {
	PK   []byte
	Host string
	// A list of possible ports
	Ports []int
	//Port the peer is currently using
	CurrentAddr *net.UDPAddr
	IsNotKnown  bool
}

//BindPort binds the client to one of the port in Ports, and returns the bound port, this should be called before all operation done on the client
func (c *Client) BindPort() (uint16, error) {
	c.nonce = crypto.RandomBytes(crypto.NonceSize)
	for _, port := range c.Ports {
		addr := net.UDPAddr{
			Port: int(port),
			IP:   net.ParseIP("0.0.0.0"),
		}
		conn, err := net.ListenUDP("udp", &addr) // code does not block here
		if err == nil {
			return port, nil
		}
		conn.SetReadBuffer(1048576)
		c.conn = conn
	}
	return 0, errors.New("Unable to bind port")
}

//StartListening start listens to all incoming packets, should be called after bind port
func (c *Client) StartListening() {
	for {
		buffer := make([]byte, 4096)
		size, addr, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error receiving data from " + addr.String())
			continue
		}
		buffer = buffer[:size]
		pk, id, plaintext, err := crypto.DecryptBytes(buffer, c.SecretKey)
		if err != nil {
			fmt.Println("Error decrypting data from " + addr.String())
			continue
		}

		sender := c.searchPeerByPk(pk)
		if sender == nil {
			sender = new(Peer)
			sender.PK = pk
			sender.CurrentAddr = addr
			sender.IsNotKnown = true
		} else {
			sender.CurrentAddr = addr
		}

		go (*c.Callback)(sender, id, plaintext)
	}
}

//SendRawPacket sends a raw packet to a peer
func (c *Client) SendRawPacket(peer *Peer, raw []byte) {
	packet := crypto.EncryptBytes(peer.PK, c.SecretKey, raw, c.nonce)
	if peer.CurrentAddr != nil {
		c.conn.WriteToUDP(packet, peer.CurrentAddr)
	} else {
		for port := range peer.Ports {
			c.conn.WriteToUDP(packet, &net.UDPAddr{
				Port: port,
				IP:   net.ParseIP(peer.Host),
			})
		}
	}
}

func (c *Client) searchPeerByPk(pk []byte) *Peer {
	for _, peer := range c.PeerList {
		if bytes.Compare(peer.PK, pk) == 0 {
			return &peer
		}
	}
	return nil
}
