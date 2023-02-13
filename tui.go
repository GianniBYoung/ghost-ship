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

type torrentInfo trans.Torrent
type errMsg struct{ err error }
type status int

var Models []tea.Model
var torrentFields = []string{"activityDate", "addedDate", "bandwidthPriority", "comment", "corruptEver", "creator", "dateCreated", "desiredAvailable", "doneDate", "downloadDir", "downloadedEver", "downloadLimit", "downloadLimited", "error", "errorString", "eta", "etaIdle", "files", "fileStats", "hashString", "haveUnchecked", "haveValid", "honorsSessionLimits", "id", "isFinished", "isPrivate", "isStalled", "leftUntilDone", "magnetLink", "manualAnnounceTime", "maxConnectedPeers", "metadataPercentComplete", "name", "peer-limit", "peers", "peersConnected", "peersFrom", "peersGettingFromUs", "peersSendingToUs", "percentDone", "pieces", "pieceCount", "pieceSize", "priorities", "queuePosition", "rateDownload", "rateUpload", "recheckProgress", "secondsDownloading", "secondsSeeding", "seedIdleLimit", "seedIdleMode", "seedRatioLimit", "seedRatioMode", "sizeWhenDone", "startDate", "status", "trackers", "trackerStats", "totalSize", "torrentFile", "uploadedEver", "uploadLimit", "uploadLimited", "uploadRatio", "wanted", "webseeds", "webseedsSendingToUs"}
var baseStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))

const (
	MainModel status = iota
	InfoView
)

func NewModel() *Model            { return &Model{} }
func (m Model) Init() tea.Cmd     { return nil }
func (m InfoModel) Init() tea.Cmd { return nil }
func (e errMsg) Error() string    { return e.err.Error() }

type TorrentTable struct {
	selected map[int]struct{} // which to-do items are selected
	table    table.Model
	torrent  trans.Torrent
	height   int
}

type Model struct {
	torrentTable TorrentTable
	state        status
	err          error
	infoModel    InfoModel
	loaded       bool
}

type InfoModel struct {
	Tabs       []string
	TabContent []string
	activeTab  int
	focused    status
	height     int
	width      int
}

func (m *Model) Next() {
	if m.state == MainModel {
		m.state = InfoView
	} else {
		m.state++
	}
}

func (m *Model) Prev() {
	if m.state == InfoView {
		m.state = MainModel
	} else {
		m.state--
	}
}

func NewInfoModel(focused status) *InfoModel {
	infoModel := &InfoModel{focused: focused}
	infoModel.Tabs = []string{"Info", "Peers", "Files"}
	infoModel.TabContent = []string{"Selected Torrent Info Will Appear Here.", "Peer Information:", "File Information"}
	infoModel.activeTab = 0
	return infoModel
}

func (m *TorrentTable) updateTable() {
	allTorrents := getAllTorrents(*TransmissionClient)
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

	myTable := table.New(table.WithColumns(columns), table.WithRows(rows), table.WithFocused(true), table.WithHeight(m.height))
	style := table.DefaultStyles()
	style.Header = style.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(false)
	style.Selected = style.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	myTable.SetStyles(style)

	m.table = myTable
}

func (m *TorrentTable) initTable(height int) {
	allTorrents := getAllTorrents(*TransmissionClient)
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

	myTable := table.New(table.WithColumns(columns), table.WithRows(rows), table.WithFocused(true), table.WithHeight(height))
	style := table.DefaultStyles()
	style.Header = style.Header.BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).BorderBottom(true).Bold(false)
	style.Selected = style.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(false)
	myTable.SetStyles(style)

	m.torrent = allTorrents[0]
	m.table = myTable

}

// takes a torrentID and returns the torrent
func getTorrentInfo(m Model, offset int) tea.Cmd {
	return func() tea.Msg {
		torrentID, _ := strconv.Atoi(m.torrentTable.table.SelectedRow()[0])
		max := len(m.torrentTable.table.Rows())
		if torrentID+offset > 0 && torrentID+offset < max {
			torrentID += offset
		}
		torrent, err := TransmissionClient.TorrentGet(context.TODO(), torrentFields, []int64{int64(torrentID)})
		if err != nil {
			return errMsg{err}
		}
		return torrentInfo(torrent[0])
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// init tabs here
		if !m.loaded {
			m.torrentTable.height = msg.Height - 25
			m.torrentTable.initTable(m.torrentTable.height)
			m.loaded = true
			return m, cmd
		}

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case torrentInfo:
		m.torrentTable.torrent = trans.Torrent(torrentInfo(msg))

	case tea.KeyMsg:

		switch msg.String() {

		case "enter":
			cmd = getTorrentInfo(m, 0)
			cmds = append(cmds, cmd)
			return m, cmd

			// Consider looking into binds: when you jump it won't update to the current torrent
		case "j":
			cmd = getTorrentInfo(m, 1)
			cmds = append(cmds, cmd)

		case "k":
			cmd = getTorrentInfo(m, -1)
			cmds = append(cmds, cmd)

		case "left", "h":
			m.Prev()

		case "m":
			TransmissionClient.TorrentSetLocation(context.TODO(), *m.torrentTable.torrent.ID, "/media/unit/ghost-ship-testing", true)
			m.torrentTable.updateTable()

		case "l":
			Models[MainModel] = m
			Models[InfoView] = createInfoModel(m.torrentTable.torrent)
			return Models[InfoView].Update(nil)

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	m.torrentTable.table, cmd = m.torrentTable.table.Update(msg)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {

	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	cursor := strconv.Itoa(int(m.torrentTable.table.Cursor()))
	torrentName := "N/A"
	var tSplit string
	if m.loaded {
		torrentName = *m.torrentTable.torrent.Name
		tSplit = m.torrentTable.table.SelectedRow()[0]
	}
	return baseStyle.Render(m.torrentTable.table.View()) + "\n" + "Cursor: " + cursor + "\n" + "Torrent: " + torrentName + "\nID: " + tSplit
}

func renderTorrentInfo(m InfoModel) string {
	doc := strings.Builder{}
	var renderedTabs []string
	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.activeTab
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
	content := lipgloss.JoinHorizontal(lipgloss.Top, windowStyle.Render(m.TabContent[m.activeTab]))
	doc.WriteString(content)
	return docStyle.Render(doc.String())

}
func getAllTorrents(transmissionClient trans.Client) []trans.Torrent {
	torrents, err := transmissionClient.TorrentGetAll(context.TODO())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
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
	activeTabBorder   = tabBorderWithBottom("┘", "─", "└")
	// figure out how to manage height and scroll
	docStyle         = lipgloss.NewStyle().Padding(1, 2, 1, 2).MaxWidth(200).MaxHeight(50)
	highlightColor   = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle   = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle      = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Left).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func createInfoModel(torrent trans.Torrent) (infoModel InfoModel) {
	infoModel.Tabs = []string{"Info", "Peers", "Files"}
	infoModel.activeTab = 0
	//0 Info 1 peers 2 files
	info := fmt.Sprintf("Name: %s \n\nStatus: %v \n\nSize: %v \n\nDate Added: %v \n\nError: %v \n\nETA: %v", *torrent.Name, parseStatus(torrent), string(torrent.TotalSize.GBString()), torrent.AddedDate.String(), *torrent.ErrorString, *torrent.Eta)

	var peers string
	p := torrent.Peers
	if len(p) > 0 {
		for _, peer := range p {
			peers += fmt.Sprintf("Peer Info:\n%s\nAddress: %s\nDownload Speed %s\nProgress: %v\nRate to Client: %v\nRate to Peer: %v\n\n", peer.ClientName, peer.Address, peer.ConvertDownloadSpeed().ByteString(), peer.Progress, peer.RateToClient, peer.RateToPeer)
		}
	} else {
		peers = "No Peers or maidens to be seen" + "\n\n"
	}

	files := fmt.Sprintf("Files:\n\n")
	f := torrent.Files

	for _, file := range f {
		files += fmt.Sprintf("%s\n", file.Name)
	}

	// files += "Files?" + *torrent.Pieces + "\n"

	infoModel.TabContent = []string{info, peers, files}
	return infoModel
}

func (m InfoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case torrentInfo:
		createInfoModel(trans.Torrent(msg))

	case tea.KeyMsg:
		switch msg.String() {

		case "right", "l":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
		case "left", "h":
			m.activeTab = max(m.activeTab-1, 0)

		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			return Models[MainModel].Update(nil)
		}

	}

	return m, cmd
}

func (m InfoModel) View() string {
	return renderTorrentInfo(m)
}
