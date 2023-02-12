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
	model Status = iota
	infoView
)

var models []tea.Model
var torrentFields = []string{"activityDate", "addedDate", "bandwidthPriority", "comment", "corruptEver", "creator", "dateCreated", "desiredAvailable", "doneDate", "downloadDir", "downloadedEver", "downloadLimit", "downloadLimited", "error", "errorString", "eta", "etaIdle", "files", "fileStats", "hashString", "haveUnchecked", "haveValid", "honorsSessionLimits", "id", "isFinished", "isPrivate", "isStalled", "leftUntilDone", "magnetLink", "manualAnnounceTime", "maxConnectedPeers", "metadataPercentComplete", "name", "peer-limit", "peers", "peersConnected", "peersFrom", "peersGettingFromUs", "peersSendingToUs", "percentDone", "pieces", "pieceCount", "pieceSize", "priorities", "queuePosition", "rateDownload", "rateUpload", "recheckProgress", "secondsDownloading", "secondsSeeding", "seedIdleLimit", "seedIdleMode", "seedRatioLimit", "seedRatioMode", "sizeWhenDone", "startDate", "status", "trackers", "trackerStats", "totalSize", "torrentFile", "uploadedEver", "uploadLimit", "uploadLimited", "uploadRatio", "wanted", "webseeds", "webseedsSendingToUs"}

var baseStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))

type TorrentInfo trans.Torrent
type ErrMsg struct{ err error }
type Status int

func NewModel() *Model            { return &Model{} }
func (m Model) Init() tea.Cmd     { return nil }
func (m InfoModel) Init() tea.Cmd { return nil }
func (e ErrMsg) Error() string    { return e.err.Error() }

type torrentTable struct {
	columns  []table.Column
	cursor   int // which to-do list item our cursor is pointing at
	rows     []table.Row
	selected map[int]struct{} // which to-do items are selected
	table    table.Model
	torrent  trans.Torrent
	torrents []trans.Torrent
}

type Model struct {
	torrentTable torrentTable
	state        Status
	err          error
	infoModel    InfoModel
	loaded       bool
}

type InfoModel struct {
	Tabs       []string
	TabContent []string
	activeTab  int
	focused    Status
}

func (m *Model) Next() {
	if m.state == model {
		m.state = infoView
	} else {
		m.state++
	}
}
func (m *Model) Prev() {
	if m.state == infoView {
		m.state = model
	} else {
		m.state--
	}
}

func NewInfoModel(focused Status) *InfoModel {
	infoModel := &InfoModel{focused: focused}
	infoModel.Tabs = []string{"Info", "Peers", "Files"}
	infoModel.TabContent = []string{"Selected Torrent Info Will Appear Here.", "Peer Information:", "File Information"}
	infoModel.activeTab = 0
	return infoModel
}

func (m *torrentTable) initTable(height int) {
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

	m.rows = rows
	m.columns = columns
	m.torrent = allTorrents[0]
	m.torrents = allTorrents
	m.table = myTable

}

// takes a torrentID and returns the torrent
func getTorrentInfo(torrentID string) tea.Cmd {
	return func() tea.Msg {
		torrentID, _ := strconv.Atoi(torrentID)
		torrent, err := TransmissionClient.TorrentGet(context.TODO(), torrentFields, []int64{int64(torrentID)})
		if err != nil {
			return ErrMsg{err}
		}
		return TorrentInfo(torrent[0])
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// init tabs here
		if !m.loaded {
			m.torrentTable.initTable(msg.Height - 5)
			m.loaded = true
			m.torrentTable.table, cmd = m.torrentTable.table.Update(msg)
			return m, cmd
		}

	case ErrMsg:
		m.err = msg
		return m, tea.Quit

	case TorrentInfo:
		m.torrentTable.torrent = trans.Torrent(TorrentInfo(msg))
		m.torrentTable.table, cmd = m.torrentTable.table.Update(msg)
		return m, cmd

	case tea.KeyMsg:

		switch msg.String() {

		case "enter":
			cmd = getTorrentInfo(m.torrentTable.table.SelectedRow()[0])
			return m, cmd

		// case "right", "l":
		// 	m.Next()

		case "left", "h":
			m.Prev()

		case "l":
			models[model] = m
			models[infoView] = createInfoModel(m.torrentTable.torrent)
			// cmd = getTorrentInfo(m.torrentTable.table.SelectedRow()[0])
			// return models[infoView].Update(cmd)
			return models[infoView].Update(nil)

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	m.torrentTable.table, cmd = m.torrentTable.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {

	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	return baseStyle.Render(m.torrentTable.table.View()) + "\n"
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
	doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).Render(m.TabContent[m.activeTab]))
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
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func createInfoModel(torrent trans.Torrent) (infoModel InfoModel) {
	infoModel.Tabs = []string{"Info", "Peers", "Files"}
	infoModel.activeTab = 0
	//0 - Info
	//1 - peers
	//2 - files
	var info string
	info += "Name: " + *torrent.Name + "\n"
	info += "Status: " + parseStatus(torrent) + "\n"
	info += "Size: " + string(torrent.TotalSize.GBString()) + "\n"
	info += "Date Added: " + torrent.AddedDate.String() + "\n"
	info += "Error: " + *torrent.ErrorString + "\n"
	info += "ETA: " + fmt.Sprint(*torrent.Eta) + "\n"

	peers := "Placeholder"
	// p := *torrent.Peers[0]
	// peers += "Peer Address: " + p.Address + "\n"
	// peers += "Download Speed: " + p.ConvertDownloadSpeed().ByteString() + "\n"
	files := "Placeholder"
	// files += "Files?" + *torrent.Pieces + "\n"
	infoModel.TabContent = []string{info, files, peers}
	return infoModel
}

func (m InfoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case TorrentInfo:
		createInfoModel(trans.Torrent(msg))

	case tea.KeyMsg:
		switch msg.String() {

		case "right", "l":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
		case "left", "h":
			m.activeTab = max(m.activeTab-1, 0)

		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m InfoModel) View() string {
	return renderTorrentInfo(m)
}

func main() {
	transmissionClientInit()
	models = []tea.Model{NewModel(), NewInfoModel(infoView)}

	m := models[model]

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
