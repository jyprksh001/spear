package ui

import (
	"encoding/base64"
	"math"
	"strconv"
	"time"

	"github.com/gdamore/tcell"
	"github.com/hexdiract/spear/core/crypto"
	"github.com/hexdiract/spear/core/network"
)

type layout struct {
	client            *network.Client
	selectedPeerIndex int
	finish            bool
}

//NewLayout creates a new CUI layout
func NewLayout(client *network.Client) {
	layout := &layout{
		client:            client,
		selectedPeerIndex: 0,
	}
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := screen.Init(); err != nil {
		panic(err)
	}
	go layout.handleEvent(&screen)
	for !layout.finish {
		layout.tick(&screen)
		screen.Show()
		time.Sleep(time.Millisecond * 50)
	}
	screen.Fini()
}

func (layout *layout) tick(screen *tcell.Screen) {
	(*screen).Clear()
	writer := &writer{screen: screen}
	writer.writeAt("  Current public key: " + base64.StdEncoding.EncodeToString(crypto.CreatePublicKey(layout.client.SecretKey)))
	writer.nextLine()
	writer.writeAt("  Up or Down arrow key to select peer.")
	writer.nextLine()
	writer.writeAt("  9 key to decrease volume of peer.")
	writer.nextLine()
	writer.writeAt("  0 key to increase volume of peer.")
	writer.nextLine()
	writer.writeAt("  Q to quit.")
	writer.nextLine()
	writer.nextLine()
	writer.x += 2
	writer.writeAt("Peer")
	writer.x += 50
	writer.writeAt("Status")
	writer.x += 15
	writer.writeAt("Volume")
	writer.x += 10
	writer.nextLine()
	for i, peer := range layout.client.PeerList {
		if i == layout.selectedPeerIndex {
			writer.writeAt(">")
		}
		writer.x += 2
		writer.writeAt(peer.DisplayName())
		writer.x += 50
		writer.writeAt(peer.Status())
		writer.x += 15
		vol := strconv.Itoa(int(math.Round(float64(peer.Volume*10)))*10) + "%"
		writer.writeAt(vol)
		writer.nextLine()
	}
}

func (layout *layout) handleEvent(screen *tcell.Screen) {
	for event := (*screen).PollEvent(); event != nil; event = (*screen).PollEvent() {
		if keyEvent, ok := event.(*tcell.EventKey); ok {
			layout.handleKey(screen, keyEvent)
		}
	}
}

func (layout *layout) handleKey(screen *tcell.Screen, event *tcell.EventKey) {
	switch event.Rune() {
	case 'q':
		layout.finish = true
	case '9':
		peer := layout.client.PeerList[layout.selectedPeerIndex]
		if peer.Volume > 0 {
			peer.Volume -= 0.1
		}
	case '0':
		peer := layout.client.PeerList[layout.selectedPeerIndex]
		if peer.Volume < 2 {
			peer.Volume += 0.1
		}
	}
	switch event.Key() {
	case tcell.KeyCtrlC:
		layout.finish = true
	case tcell.KeyUp:
		layout.selectedPeerIndex++
	case tcell.KeyDown:
		layout.selectedPeerIndex--
	}

	m := len(layout.client.PeerList)
	layout.selectedPeerIndex = (layout.selectedPeerIndex + m) % m
}
