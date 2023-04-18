package main

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	trans "github.com/hekmon/transmissionrpc/v2"
)

func (m TextInputModel) Init() tea.Cmd { return textinput.Blink }

type torrentMoved bool

type TextInputModel struct {
	textInput        textinput.Model
	torrent          trans.Torrent
	selectedTorrents map[int]trans.Torrent
}

func NewTextInputModel(torrentID int, torrents torrentSelected) TextInputModel {
	textInputModel := textinput.New()
	textInputModel.Placeholder = "Enter New Path"
	textInputModel.Focus()
	torrent, _ := TransmissionClient.TorrentGet(context.TODO(), torrentFields, []int64{int64(torrentID)})
	return TextInputModel{textInput: textInputModel, torrent: torrent[0], selectedTorrents: torrents}
}

func (m TextInputModel) moveTorrents() tea.Cmd {
	return func() tea.Msg {
		for key := range m.selectedTorrents {
			MoveTorrent(*m.selectedTorrents[key].ID, m.textInput.Value())
		}
		return torrentMoved(true)
	}
}

// func (m TextInputModel) grabTextArea() tea.Msg {
// 	m.textValue = m.textInput.Value()
// 	return textGrabbed(true)
// }

func (m TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case torrentMoved:
		Models[TextInputView] = nil
		return Models[MainModel].Update(callBackMsg(true))

	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			cmd = m.moveTorrents()
			m.textInput.Blur()
			return m, cmd
		}

	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m TextInputModel) View() string {
	var torrents string
	for key := range m.selectedTorrents {
		torrents += *m.selectedTorrents[key].Name + "\n"
	}

	return fmt.Sprintf(
		`The Selected torrent is %s(ID:%d) located at %s
Enter New Location:
%s
The following torrents will be moved to: %s
%s

`,
		*m.torrent.Name,
		*m.torrent.ID,
		*m.torrent.DownloadDir,
		m.textInput.View(),
		m.textInput.Value(),
		torrents,
	)
}
