package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	trans "github.com/hekmon/transmissionrpc/v2"
)

func main() {
	transmissionClientInit()
	Models = []tea.Model{NewModel(), NewInfoModel(InfoView), NewTextInputModel(1, make(map[int]trans.Torrent))}

	m := Models[MainModel]

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error starting bubbletea: %v", err)
		os.Exit(1)
	}
}
