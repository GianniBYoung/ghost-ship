package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m TextInputModel) Init() tea.Cmd { return textinput.Blink }

type TextInputModel struct {
	textInput textinput.Model
	focused   status
	err       error
}

func NewTextInputModel(focused status) TextInputModel {
	textInputModel := textinput.New()
	textInputModel.Placeholder = "Set Torrent Location"
	textInputModel.Focus()
	return TextInputModel{textInput: textInputModel, err: nil}
}

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

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m TextInputModel) View() string {
	return fmt.Sprintf(
		"Enter New Location\n\n%s\n\n%s",
		m.textInput.View(),
		"The following torrents will be moved to `"+m.textInput.Value(),
	) + "`:\n"
}
