package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	trans "github.com/hekmon/transmissionrpc/v2"
)

var transmissionPassword = os.Getenv("TRANSMISSIONPASSWORD")
var transmissionUserName = os.Getenv("TRANSMISSIONUSERNAME")
var transmissionIP = os.Getenv("TRANSMISSIONIP")
var transmissionClient, err = trans.New(transmissionIP, transmissionUserName, transmissionPassword, nil)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type torrentInfo trans.Torrent

type model struct {
	torrents    []trans.Torrent
	cursor      int              // which to-do list item our cursor is pointing at
	selected    map[int]struct{} // which to-do items are selected
	torrentName string
	torrent     trans.Torrent
	columns     []table.Column
	rows        []table.Row
	table       table.Model
}

func initialModel() model {
	allTorrents := getAllTorrents(*transmissionClient)
	var rows []table.Row

	for _, torrent := range allTorrents {
		rows = append(rows, buildRow(torrent))
	}

	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Name", Width: 45},
		{Title: "Status", Width: 23},
		{Title: "Location", Width: 35},
	}

	table := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(30),
	)
	// style := table.DefaultStyles()
	// style.Header = style.Header.
	// 	BorderStyle(lipgloss.NormalBorder()).
	// 	BorderForeground(lipgloss.Color("240")).
	// 	BorderBottom(true).
	// 	Bold(false)
	// style.Selected = style.Selected.
	// 	Foreground(lipgloss.Color("229")).
	// 	Background(lipgloss.Color("57")).
	// 	Bold(false)
	// table.SetStyles(style)

	return model{
		torrents: allTorrents,
		torrent:  allTorrents[0],
		selected: make(map[int]struct{}),
		rows:     rows,
		columns:  columns,
		table:    table,
	}
}

func (m model) Init() tea.Cmd { return nil }

// takes a torrent and returns the torrent
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

		// What key was pressed?
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		// Move the cursor
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				return m, teaTorrentInfo(m.torrents[m.cursor])
			}

		case "down", "j":
			if m.cursor < len(m.torrents)-1 {
				m.cursor++
				return m, teaTorrentInfo(m.torrents[m.cursor])
			}

		// Toggle 'selected' state
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

	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	s := "Transmission Torrents\n\n"
	return baseStyle.Render(m.table.View()) + "\n"

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

func buildRow(torrent trans.Torrent) table.Row {
	torrentName := string(*torrent.Name)
	torrentID := strconv.Itoa(int(*torrent.ID))
	var torrentStatus string

	switch *torrent.Status {
	case 0:
		torrentStatus = "Stopped"
	case 1:
		torrentStatus = "Checking Files"
	case 2:
		torrentStatus = "Files Checked"
	case 3:
		torrentStatus = "Queued for Download"
	case 4:
		torrentStatus = "Downloading"
	case 5:
		torrentStatus = "Waiting for Seeds"
	case 6:
		torrentStatus = "Actively Seeding"
	case 7:
		torrentStatus = "No Peers Found"
	}

	return table.Row{torrentID, torrentName, torrentStatus, *torrent.DownloadDir}

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
