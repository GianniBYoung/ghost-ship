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
var torrentFields = []string{"activityDate", "addedDate", "bandwidthPriority", "comment", "corruptEver", "creator", "dateCreated", "desiredAvailable", "doneDate", "downloadDir", "downloadedEver", "downloadLimit", "downloadLimited", "error", "errorString", "eta", "etaIdle", "files", "fileStats", "hashString", "haveUnchecked", "haveValid", "honorsSessionLimits", "id", "isFinished", "isPrivate", "isStalled", "leftUntilDone", "magnetLink", "manualAnnounceTime", "maxConnectedPeers", "metadataPercentComplete", "name", "peer-limit", "peers", "peersConnected", "peersFrom", "peersGettingFromUs", "peersSendingToUs", "percentDone", "pieces", "pieceCount", "pieceSize", "priorities", "queuePosition", "rateDownload", "rateUpload", "recheckProgress", "secondsDownloading", "secondsSeeding", "seedIdleLimit", "seedIdleMode", "seedRatioLimit", "seedRatioMode", "sizeWhenDone", "startDate", "status", "trackers", "trackerStats", "totalSize", "torrentFile", "uploadedEver", "uploadLimit", "uploadLimited", "uploadRatio", "wanted", "webseeds", "webseedsSendingToUs"}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type torrentInfo trans.Torrent
type errMsg struct{ err error }

func (m model) Init() tea.Cmd  { return nil }
func (e errMsg) Error() string { return e.err.Error() }

type model struct {
	torrents    []trans.Torrent
	cursor      int              // which to-do list item our cursor is pointing at
	selected    map[int]struct{} // which to-do items are selected
	torrentName string
	torrent     trans.Torrent
	columns     []table.Column
	rows        []table.Row
	table       table.Model
	err         error
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
		{Title: "Status", Width: 20},
		{Title: "Size", Width: 8},
		{Title: "Location", Width: 35},
	}

	myTable := table.New(table.WithColumns(columns), table.WithRows(rows), table.WithFocused(true), table.WithHeight(60))
	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	style.Selected = style.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	myTable.SetStyles(style)

	return model{
		torrents: allTorrents,
		torrent:  allTorrents[0],
		selected: make(map[int]struct{}),
		rows:     rows,
		columns:  columns,
		table:    myTable,
	}
}

// takes a torrentID and returns the torrent
func getTorrentInfo(torrentID string) tea.Cmd {
	return func() tea.Msg {
		torrentID, _ := strconv.Atoi(torrentID)
		torrent, err := transmissionClient.TorrentGet(context.TODO(), torrentFields, []int64{int64(torrentID)})
		if err != nil {
			return errMsg{err}
		}

		return torrentInfo(torrent[0])
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case torrentInfo:
		m.torrent = trans.Torrent(torrentInfo(msg))
		m.table, cmd = m.table.Update(msg)
		return m, cmd

	// Is it a key press?
	case tea.KeyMsg:

		// What key was pressed?
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}

		case "enter":
			return m, getTorrentInfo(m.table.SelectedRow()[0])

		case "ctrl+c", "q":
			return m, tea.Quit

		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	s := baseStyle.Render(m.table.View()) + "\n"
	if *m.torrent.Name != "" {
		s += fmt.Sprintf("\n\nThe selected Torrent is: %s ", *m.torrent.Name)
	} else {
		s += fmt.Sprintf("\n\nThe torrent is nil: %v", m)
	}
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

func parseStatus(torrent trans.Torrent) string {
	switch *torrent.Status {
	case 0:
		return "Stopped"
	case 1:
		return "Checking Files"
	case 2:
		return "Files Checked"
	case 3:
		return "Queued for Download"
	case 4:
		return "Downloading"
	case 5:
		return "Waiting for Seeds"
	case 6:
		return "Seeding"
	case 7:
		return "No Peers Found"
	}
	return "Status unknown???"

}

func buildRow(torrent trans.Torrent) table.Row {
	torrentID := strconv.Itoa(int(*torrent.ID))
	torrentStatus := parseStatus(torrent)
	// torrentSize := *torrent.TotalSize
	torrentSize := string(torrent.TotalSize.GBString())
	return table.Row{torrentID, string(*torrent.Name), torrentStatus, torrentSize, *torrent.DownloadDir}

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

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
