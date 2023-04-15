package main

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	trans "github.com/hekmon/transmissionrpc/v2"
)

func (m TextInputModel) Init() tea.Cmd { return textinput.Blink }

type TextInputModel struct {
	textInput textinput.Model
	focused   status
	torrent   trans.Torrent
	err       error
}

func NewTextInputModel(focused status, torrentID int) TextInputModel {
	textInputModel := textinput.New()
	textInputModel.Placeholder = "Enter New Path"
	textInputModel.Focus()
	torrent, _ := TransmissionClient.TorrentGet(context.TODO(), torrentFields, []int64{int64(torrentID)})
	return TextInputModel{textInput: textInputModel, err: nil, torrent: torrent[0]}
}

//TransmissionClient.TorrentSetLocation(context.TODO(), *m.torrentTable.torrent.ID, "", true)
func (m TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			return m, cmd
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m TextInputModel) View() string {
	return fmt.Sprintf(
		`The Selected torrent is %s located at %s
Enter New Location:
%s
The following torrents will be moved to "%s`,
		*m.torrent.Name,
		*m.torrent.DownloadDir,
		m.textInput.View(),
		m.textInput.Value(),
	) + "\":\n"
}
