package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	trans "github.com/hekmon/transmissionrpc/v2"
)

var transmissionPassword = os.Getenv("TRANSMISSIONPASSWORD")
var transmissionUserName = os.Getenv("TRANSMISSIONUSERNAME")
var transmissionIP = os.Getenv("TRANSMISSIONIP")

// Capitalize thie and make tranmissionClient exportable?
var transmissionClient, err = trans.New(transmissionIP, transmissionUserName, transmissionPassword, nil)

type torrentInfo trans.Torrent

type model struct {
	torrents    []trans.Torrent
	cursor      int              // which to-do list item our cursor is pointing at
	selected    map[int]struct{} // which to-do items are selected
	torrentName string
	torrent     trans.Torrent
}

func initialModel() model {
	allTorrents := getAllTorrents(*transmissionClient)
	return model{
		torrents: allTorrents,
		torrent:  allTorrents[0],
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func teaTorrentInfo(torrent trans.Torrent) tea.Cmd {
	return func() tea.Msg {
		return torrentInfo(torrent)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case torrentInfo:
		m.torrent = trans.Torrent(msg)
		return m, nil

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.torrents)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
			return m, teaTorrentInfo(m.torrents[m.cursor])
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	s := "Transmission Torrents\n\n"

	// for i, torrent := range m.torrents {
	for i := 0; i < 10; i++ {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		// s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, *torrent.Name)
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, *m.torrents[i].Name)
	}

	if *m.torrent.Name != "" {

		s += fmt.Sprintf("\n\nThe currently selected torrent is %s\n", *m.torrent.Name)
		s += fmt.Sprintf("It's ID is %v\n", *m.torrent.ID)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func getAllTorrents(transmissionClient trans.Client) []trans.Torrent {
	torrents, err := transmissionClient.TorrentGetAll(context.TODO())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		return torrents
	}
	return torrents
}

func listAllTorrents(torrents []trans.Torrent) {
	for _, torrent := range torrents {
		torrentName := *torrent.Name
		torrentID := *torrent.ID
		fmt.Println(torrentName)
		fmt.Println(torrentID)
	}
}

func main() {
	if err != nil {
		fmt.Println("Unable to create transmission client.")
		panic(err)
	}

	ok, serverVersion, serverMinimumVersion, err := transmissionClient.RPCVersion(context.TODO())
	if err != nil {
		panic(err)
	}
	if !ok {
		panic(fmt.Sprintf("Remote transmission RPC version (v%d) is incompatible with the transmission library (v%d): remote needs at least v%d",
			serverVersion, trans.RPCVersion, serverMinimumVersion))
	}
	fmt.Printf("Remote transmission RPC version (v%d) is compatible with our trans library (v%d)\n", serverVersion, trans.RPCVersion)

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	// torrents := getAllTorrents(*transmissionClient)

	// fmt.Println(*torrents[0].Name)
	// listAllTorrents(torrents)

}
