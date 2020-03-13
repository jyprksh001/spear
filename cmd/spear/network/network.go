package network

import (
	"errors"
	"log"
	"net"
	"time"

	"../audio"
	"../crypto"
)

const minimumPacketBufferSize = 5
const maximumPacketBufferSize = 10
const maximumTimeDifference = 500

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

	PeerList []*Peer

	Addr Addr
	conn *net.UDPConn
}

//Initialize setup the client, should be called first
func (client *Client) Initialize() error {
	if len(client.Addr.Candidates) == 0 {
		panic("Address candidates is empty")
	}
	conn, err := client.Addr.Bind()
	if err != nil {
		return err
	}
	conn.SetReadBuffer(0x100000)
	client.conn = conn
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

		for _, peer := range client.GetPeerByAddr(addr) {
			id, plaintext, err := crypto.DecryptBytes(buffer[:size], peer.PublicKey, client.SecretKey)
			if err == nil {
				peer.receivePacket(&Packet{
					ID:           id,
					RawData:      plaintext,
					ReceivedTime: time.Now().UTC().UnixNano() / 1000000,
				})
				break
			}
		}

	}
	done <- true
}

//SendAudioData takes raw MONO audio data and send it to another peer
func (client *Client) SendAudioData(peer *Peer, pcm []float32) {
	if len(pcm) != audio.FrameSize {
		panic("pcm size not equal to audio.FrameSize")
	}
	data := audio.CompressAudio(pcm)

	client.sendBytes(peer, append([]byte{0}, data...))
}

//sendBytes send a packet to another peer, bytes should be unencrypted.
func (client *Client) sendBytes(peer *Peer, bytes []byte) {
	ciphertext := crypto.EncryptBytes(peer.PublicKey, client.SecretKey, bytes, peer.packetID)
	peer.Addr.Write(client.conn, ciphertext)
	peer.packetID++
}

//GetPeerByAddr find a peers with the corresponding address
func (client *Client) GetPeerByAddr(addr *net.UDPAddr) []*Peer {
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
