package ui

import (
	"github.com/gdamore/tcell"
)

type writer struct {
	screen *tcell.Screen
	x, y   int
}

func (writer *writer) writeAt(line string) {
	for i := 0; i < len(line); i++ {
		(*writer.screen).SetContent(writer.x+i, writer.y, rune(line[i]), nil, 0)
	}
}

func (writer *writer) nextLine() {
	writer.y++
	writer.x = 0
}
