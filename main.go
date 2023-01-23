package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	trans "github.com/hekmon/transmissionrpc/v2"
)

const (
	mainView sessionState = iota
	infoView
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
type torrentInfoTab infoModel
type errMsg struct{ err error }
type sessionState uint

func (m mainModel) Init() tea.Cmd { return nil }
func (e errMsg) Error() string    { return e.err.Error() }

type mainModel struct {
	state       sessionState
	torrents    []trans.Torrent
	cursor      int              // which to-do list item our cursor is pointing at
	selected    map[int]struct{} // which to-do items are selected
	torrentName string
	torrent     trans.Torrent
	columns     []table.Column
	rows        []table.Row
	table       table.Model
	err         error
	infoModel   infoModel
}

type infoModel struct {
	Tabs       []string
	TabContent []string
	activeTab  int
}

func updateTorrentInfoTabs(torrent trans.Torrent) (info, peers, files string) {

	info += "Name: " + *torrent.Name + "\n"
	info += "Status: " + parseStatus(torrent) + "\n"
	info += "Size: " + string(torrent.TotalSize.GBString()) + "\n"
	// info += *allTorrents[0] + "\n"
	info += "Date Added: " + torrent.AddedDate.String() + "\n"
	info += "Error: " + *torrent.ErrorString + "\n"
	info += "ETA: " + fmt.Sprint(*torrent.Eta) + "\n"

	peers += "Placeholder"
	// p := *torrent.Peers[0]
	// peers += "Peer Address: " + p.Address + "\n"
	// peers += "Download Speed: " + p.ConvertDownloadSpeed().ByteString() + "\n"

	files += "Placeholder"
	// files += "Files?" + *torrent.Pieces + "\n"
	return info, peers, files
}

// I WILL need to make a tea.cmd to update the model
func initialModel() mainModel {
	allTorrents := getAllTorrents(*transmissionClient)
	var rows []table.Row
	var info string

	for _, torrent := range allTorrents {
		rows = append(rows, buildRow(torrent))
	}

	info, peers, files := updateTorrentInfoTabs(allTorrents[0])

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

	return mainModel{
		torrents: allTorrents,
		torrent:  allTorrents[0],
		selected: make(map[int]struct{}),
		rows:     rows,
		columns:  columns,
		table:    myTable,
		state:    mainView,
		infoModel: infoModel{
			Tabs:       []string{"Info", "Peers", "Files"},
			TabContent: []string{info, peers, files},
			activeTab:  0,
		},
	}
}

func (m mainModel) currentFocusedModel() string {
	if m.state == infoView {
		return "infoView"
	}
	return "mainView"
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

//create struct builder for this
func updateTabs(torrent trans.Torrent) tea.Cmd {
	return func() tea.Msg {

		info, peers, files := updateTorrentInfoTabs(torrent)
		infoModel := infoModel{
			Tabs:       []string{"Info", "Peers", "Files"},
			TabContent: []string{info, peers, files},
			activeTab:  0,
		}
		if err != nil {
			return errMsg{err}
		}

		return torrentInfoTab(infoModel)
	}
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case torrentInfo:
		m.torrent = trans.Torrent(torrentInfo(msg))
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)

	case torrentInfoTab:
		m.infoModel = infoModel(msg)

	case tea.KeyMsg:

		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}

		case "tab":
			if m.state == mainView {
				m.state = infoView
				cmd = updateTabs(m.torrent)
				cmds = append(cmds, cmd)
			} else {
				m.state = mainView
			}

		case "enter":
			cmd = getTorrentInfo(m.table.SelectedRow()[0])
			cmds = append(cmds, cmd)

		case "ctrl+c", "q":
			return m, tea.Quit

		case "right", "l":
			if m.state == mainView {
				m.state = infoView
			} else {
				m.infoModel.activeTab = min(m.infoModel.activeTab+1, len(m.infoModel.Tabs)-1)
			}
		case "left", "h":
			m.infoModel.activeTab = max(m.infoModel.activeTab-1, 0)
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, tea.Batch(cmds...)
}

func renderTorrentInfo(m mainModel) string {
	doc := strings.Builder{}
	var renderedTabs []string
	for i, t := range m.infoModel.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.infoModel.Tabs)-1, i == m.infoModel.activeTab
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(m.infoModel.TabContent[m.infoModel.activeTab]))
	return docStyle.Render(doc.String())

}

func (m mainModel) View() string {

	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	if m.state == infoView {

		return renderTorrentInfo(m)
		// s += fmt.Sprintf("\n\nThe selected Torrent is: %s ", *m.torrent.Name)
	} else {
		return baseStyle.Render(m.table.View()) + "\n"
	}
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

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
