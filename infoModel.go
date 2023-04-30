package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	trans "github.com/hekmon/transmissionrpc/v2"
)

func (m InfoModel) Init() tea.Cmd { return nil }

type InfoModel struct {
	Tabs       []string
	TabContent []string
	activeTab  int
	focused    status
	height     int
	width      int
}

func NewInfoModel(focused status) *InfoModel {
	infoModel := &InfoModel{focused: focused}
	infoModel.Tabs = []string{"Info", "Peers", "Files"}
	infoModel.TabContent = []string{"Selected Torrent Info Will Appear Here.", "Peer Information:", "File Information"}
	infoModel.activeTab = 0
	return infoModel
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
