package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	trans "github.com/hekmon/transmissionrpc/v2"
)

type torrentInfo trans.Torrent
type torrentSelected map[int]trans.Torrent
type errMsg struct{ err error }
type callBackMsg bool
type status int

var Models []tea.Model
var torrentFields = []string{"activityDate", "addedDate", "bandwidthPriority", "comment", "corruptEver", "creator", "dateCreated", "desiredAvailable", "doneDate", "downloadDir", "downloadedEver", "downloadLimit", "downloadLimited", "error", "errorString", "eta", "etaIdle", "files", "fileStats", "hashString", "haveUnchecked", "haveValid", "honorsSessionLimits", "id", "isFinished", "isPrivate", "isStalled", "leftUntilDone", "magnetLink", "manualAnnounceTime", "maxConnectedPeers", "metadataPercentComplete", "name", "peer-limit", "peers", "peersConnected", "peersFrom", "peersGettingFromUs", "peersSendingToUs", "percentDone", "pieces", "pieceCount", "pieceSize", "priorities", "queuePosition", "rateDownload", "rateUpload", "recheckProgress", "secondsDownloading", "secondsSeeding", "seedIdleLimit", "seedIdleMode", "seedRatioLimit", "seedRatioMode", "sizeWhenDone", "startDate", "status", "trackers", "trackerStats", "totalSize", "torrentFile", "uploadedEver", "uploadLimit", "uploadLimited", "uploadRatio", "wanted", "webseeds", "webseedsSendingToUs"}
var baseStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240"))

const (
	MainModel status = iota
	InfoView
	TextInputView
)

func NewModel() *Model         { return &Model{} }
func (m Model) Init() tea.Cmd  { return nil }
func (e errMsg) Error() string { return e.err.Error() }

type TorrentTable struct {
	selectedTorrents map[int]trans.Torrent
	table            table.Model
	torrent          trans.Torrent
	height           int
	width            int
}

type Model struct {
	torrentTable TorrentTable
	state        status
	err          error
	infoModel    InfoModel
	loaded       bool
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

func buildRow(torrent trans.Torrent, headers []string) table.Row {
	var row table.Row

	for _, key := range headers {
		switch key {
		case "ID":
			row = append(row, strconv.Itoa(int(*torrent.ID)))

		case "Name":
			row = append(row, *torrent.Name)

		case "Status":
			row = append(row, parseStatus(torrent))

		case "Size":
			row = append(row, string(torrent.TotalSize.GBString()))

		case "Ratio":
			row = append(row, fmt.Sprintf("%.2f", *torrent.UploadRatio))

		case "Location":
			row = append(row, *torrent.DownloadDir)

		case "Activity Date":
			date := *torrent.ActivityDate
			row = append(row, date.Format("2006-01-02"))

		case "Download Rate":
			row = append(row, strconv.Itoa(int(*torrent.RateDownload)))

		case "Upload Rate":
			row = append(row, strconv.Itoa(int(*torrent.RateUpload)))

		case "Uploaded Ever":
			row = append(row, strconv.Itoa(int(*torrent.UploadedEver)))

		case "Error":
			row = append(row, *torrent.ErrorString)

		// deal with this slice
		case "Lables":

			var labels string
			for _, label := range torrent.Labels {
				labels += label + ","
			}
			row = append(row, labels)

		// deal with this slice
		case "Trackers":
			var trackers string
			for _, tracker := range torrent.TrackerStats {
				trackers += tracker.Host + ","
			}
			row = append(row, trackers)

		}

	}
	return row
}

func SetColumns(t TorrentTable) (columns []table.Column, headers []string) {

	totalColumns := 1
	for range Cfg.UI.Columns {
		totalColumns++
	}
	offset := 0

	for _, c := range Cfg.UI.Columns {
		switch c {
		case "ID":
			offset += 4
		case "Ratio":
			offset += 5

		case "Size":
			offset += 10

		case "Status":
			offset += 20
		}
	}

	maxColumnSize := (t.width + offset) / totalColumns

	for _, c := range Cfg.UI.Columns {
		var column table.Column

		switch c {

		case "ID":
			column = table.Column{Title: c, Width: 4}

		case "Ratio":
			column = table.Column{Title: c, Width: 5}

		case "Size":
			column = table.Column{Title: c, Width: 10}

		case "Status":
			column = table.Column{Title: c, Width: 20}

		default:
			column = table.Column{Title: c, Width: maxColumnSize}

		}
		columns = append(columns, column)
		headers = append(headers, c)
	}

	return columns, headers
}

func (m *TorrentTable) updateTable() {
	allTorrents := getAllTorrents(*TransmissionClient)

	visibleColumns, headers := SetColumns(*m)
	var rows []table.Row
	for _, torrent := range allTorrents {
		rows = append(rows, buildRow(torrent, headers))
	}

	myTable := table.New(table.WithColumns(visibleColumns), table.WithRows(rows), table.WithFocused(true), table.WithHeight(m.height), table.WithWidth(m.width))
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

// adds a torrent to the selcted list
func selectTorrent(m Model) tea.Cmd {
	return func() tea.Msg {
		offset := 1
		cursor := m.torrentTable.table.Cursor()
		_, exists := m.torrentTable.selectedTorrents[cursor]
		if exists {
			delete(m.torrentTable.selectedTorrents, cursor)
		} else {
			torrent, _ := TransmissionClient.TorrentGet(context.TODO(), torrentFields, []int64{int64(cursor + offset)})
			m.torrentTable.selectedTorrents[m.torrentTable.table.Cursor()] = torrent[0]
		}

		return torrentSelected(m.torrentTable.selectedTorrents)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.loaded {
			m.torrentTable.height = msg.Height - 25
			m.torrentTable.width = msg.Width - 5
			m.torrentTable.updateTable()
			m.torrentTable.selectedTorrents = make(map[int]trans.Torrent)
			m.loaded = true
			return m, cmd
		}

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case torrentInfo:
		m.torrentTable.torrent = trans.Torrent(torrentInfo(msg))

	case torrentSelected:
		m.torrentTable.selectedTorrents = msg

	case callBackMsg:
		m.torrentTable.updateTable()

	case tea.KeyMsg:

		switch msg.String() {

		case "enter":
			cmd = selectTorrent(m)
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

		case "l":
			Models[MainModel] = m
			Models[InfoView] = createInfoModel(m.torrentTable.torrent)
			return Models[InfoView].Update(nil)

		case "M", "m":
			cmd = getTorrentInfo(m, 0)
			Models[MainModel] = m
			Models[TextInputView] = NewTextInputModel(int(*m.torrentTable.torrent.ID), m.torrentTable.selectedTorrents)
			return Models[TextInputView].Update(nil)

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
	var selectedTorrents string
	if m.loaded {
		torrentName = *m.torrentTable.torrent.Name
		tSplit = m.torrentTable.table.SelectedRow()[0]
		var selected []string
		for _, element := range m.torrentTable.selectedTorrents {
			selected = append(selected, *element.Name)
		}
		sort.Strings(selected)
		selectedTorrents = fmt.Sprintln(strings.Join(selected, "\n"))

	}
	return baseStyle.Render(m.torrentTable.table.View()) + "\n" + "Cursor: " + cursor + "\n" + "Torrent: " + torrentName + "\nID: " + tSplit + "\n" + selectedTorrents
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
