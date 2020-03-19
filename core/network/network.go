package network

import (
	"errors"
	"log"
	"net"
	"time"

	"github.com/hexdiract/spear/core/crypto"
)

//DeterminableAddr is a container of candidates and the current one
type DeterminableAddr struct {
	current    *net.UDPAddr
	Candidates []*net.UDPAddr
}

//Client refers to the backend of the client containing all basic information needed by the core
type Client struct {
	SecretKey []byte

	PeerList []*Peer

	Addr DeterminableAddr
	conn *net.UDPConn
}

//Initialize setup the client, should be called first
func (client *Client) Initialize() error {
	if len(client.Addr.Candidates) == 0 {
		panic("Address candidates is empty")
	}
	conn, err := client.bind()
	if err != nil {
		return err
	}
	conn.SetReadBuffer(0x100000)
	client.conn = conn
	for _, p := range client.PeerList {
		p.init(client)
	}

	go client.start()
	return nil
}

func (client *Client) start() {
	for {
		buffer := make([]byte, 0x1000)
		size, addr, err := client.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, peer := range client.getPeerByAddr(addr) {
			id, plaintext, err := crypto.DecryptBytes(buffer[:size], peer.PublicKey, client.SecretKey)
			if err == nil {
				packet := &Packet{
					ID:           id,
					RawData:      plaintext,
					ReceivedTime: time.Now().UTC().UnixNano() / 1000000,
				}
				switch plaintext[0] {
				case AudioID:
					peer.receiveAudioPacket(packet)
				default:
					log.Printf("Unsupported data %d\n", plaintext[0])
				}
				break
			}
		}

	}
}

func (client *Client) getPeerByAddr(addr *net.UDPAddr) []*Peer {
	peers := []*Peer{}
	for _, peer := range client.PeerList {
		for _, paddr := range peer.Addr.Candidates {
			if paddr.IP.Equal(addr.IP) && paddr.Port == addr.Port {
				peers = append(peers, peer)
			}
		}
	}
	return peers
}

func (client *Client) bind() (*net.UDPConn, error) {
	for _, cand := range client.Addr.Candidates {
		conn, err := net.ListenUDP("udp", cand)
		if err == nil {
			client.Addr.current = cand
			log.Println("Successfully bound to", cand.String())
			return conn, nil
		}
		log.Println("Attempted to bind to", cand.String())
	}
	return nil, errors.New("Unable to bind to any candidates")
}

func (client *Client) writeTo(peer *Peer, data []byte) {
	if peer.Addr.current == nil {
		for _, cand := range peer.Addr.Candidates {
			client.conn.WriteToUDP(data, cand)
		}
	} else {
		client.conn.WriteToUDP(data, peer.Addr.current)
	}
}
