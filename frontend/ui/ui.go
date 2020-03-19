package ui

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"

	"github.com/hexdiract/spear/core/crypto"
	"github.com/hexdiract/spear/core/network"
	"github.com/jroimartin/gocui"
)

var (
	infoSizes = []int{50, 15, 10}
	totalSize = float32(sum(infoSizes...))
)

type layout struct {
	client            *network.Client
	selectedPeerIndex int
}

//NewLayout creates a new CUI layout
func NewLayout(client *network.Client) {
	layout := &layout{
		client: client,
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout.updateLayout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, moveIndex(layout, 1)); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, moveIndex(layout, -1)); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", '0', gocui.ModNone, changeVol(layout, 0.1)); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", '9', gocui.ModNone, changeVol(layout, -0.1)); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

	g.Update(layout.updateLayout)
}

func (layout *layout) updateLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	maxX--
	if _, err := g.SetView("main", 0, 0, maxX, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		layout.printAll(g)
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func moveIndex(layout *layout, offset int) func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		layout.selectedPeerIndex += offset
		layout.selectedPeerIndex %= len(layout.client.PeerList)
		if layout.selectedPeerIndex < 0 {
			layout.selectedPeerIndex += len(layout.client.PeerList)
		}
		layout.printAll(gui)
		return nil
	}
}

func changeVol(layout *layout, offset float32) func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		peer := layout.client.PeerList[layout.selectedPeerIndex]
		peer.Volume += offset
		if peer.Volume > 2 {
			peer.Volume = 2
		}
		if peer.Volume < 0 {
			peer.Volume = 0
		}
		layout.printAll(gui)
		return nil
	}
}

func (layout *layout) printAll(gui *gocui.Gui) {
	view, _ := gui.View("main")
	view.Clear()
	fmt.Fprintln(view, "  Current public key: "+base64.StdEncoding.EncodeToString(crypto.CreatePublicKey(layout.client.SecretKey)))
	fmt.Fprintln(view, "  Up or Down arrow key to select peer.")
	fmt.Fprintln(view, "  9 key to decrease volume of peer.")
	fmt.Fprintln(view, "  0 key to increase volume of peer.")
	fmt.Fprintln(view, "  Q to quit.")
	fmt.Fprintln(view)
	maxX, _ := gui.Size()
	printPeer([]string{"  Peer ID", "Status", "Vol."}, maxX, view)
	for i, peer := range layout.client.PeerList {
		selected := "  "
		if i == layout.selectedPeerIndex {
			selected = "> "
		}
		vol := strconv.Itoa(int(peer.Volume*100)) + "%"
		printPeer([]string{selected + peer.DisplayName(), peer.Status(), vol}, maxX, view)
	}
}

func printPeer(info []string, width int, view *gocui.View) {
	pattern := []string{}
	for _, v := range infoSizes {
		w := int(float32(v) / totalSize * float32(width))
		pattern = append(pattern, "%-"+strconv.Itoa(w)+"s")
	}
	for i, v := range info {
		fmt.Fprintf(view, pattern[i], v)
	}
	fmt.Fprintln(view)
}

func sum(input ...int) int {
	sum := 0

	for _, i := range input {
		sum += i
	}

	return sum
}
