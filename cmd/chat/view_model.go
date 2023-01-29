package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/wisepythagoras/owncom/core"
	"github.com/wisepythagoras/owncom/crypto"
)

type (
	errMsg error
)

type UserMessage struct {
	From          string
	Message       string
	LabelRenderer func(str string) string
}

type ViewModel struct {
	viewport         viewport.Model
	messages         []UserMessage
	messageInput     textarea.Model
	currentUserStyle lipgloss.Style
	otherUserStyle   lipgloss.Style
	handler          *core.Handler
	aesGcmKey        *crypto.AESGCMKey
	err              error
	screenWidth      int
}

func (m ViewModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ViewModel) broadCastMessage(userMsg UserMessage) {
	var packets []core.Packet
	var err error

	msg := core.Message{
		Msg:    []byte(userMsg.Message),
		Module: m.handler.Module,
	}

	if m.aesGcmKey != nil {
		packets, err = msg.PacketsAESGCM(m.aesGcmKey)

		if err != nil {
			fmt.Println(err)
		}

		m.handler.Send(packets)
	}
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		messageInputCmd tea.Cmd
		viewportCmd     tea.Cmd
	)

	m.messageInput, messageInputCmd = m.messageInput.Update(msg)
	m.viewport, viewportCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport = viewport.New(msg.Width, msg.Height-7)
		m.messageInput.SetWidth(msg.Width)
		m.screenWidth = msg.Width

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			newMessage := UserMessage{
				From:          "You: ",
				Message:       m.messageInput.Value(),
				LabelRenderer: m.currentUserStyle.Render,
			}

			// Broadcast the message
			m.broadCastMessage(newMessage)

			m.messages = append(m.messages, newMessage)
			messageContent := make([]string, 0)

			for _, msg := range m.messages {
				someMessage := msg.LabelRenderer(msg.From) + msg.Message
				messageContent = append(messageContent, someMessage)
			}

			m.viewport.SetContent(strings.Join(messageContent, "\n"))
			m.messageInput.Reset()
			m.viewport.GotoBottom()
		}

	case UserMessage:
		if msg.LabelRenderer == nil {
			msg.LabelRenderer = m.otherUserStyle.Render
		}

		m.messages = append(m.messages, msg)
		messageContent := make([]string, 0)

		for _, msg := range m.messages {
			someMessage := msg.LabelRenderer(msg.From) + msg.Message
			messageContent = append(messageContent, someMessage)
		}

		m.viewport.SetContent(lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(messageContent, "\n")))
		m.messageInput.Reset()
		m.viewport.GotoBottom()

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(messageInputCmd, viewportCmd)
}

func (m ViewModel) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		viewportStyle(m.screenWidth).Render(m.viewport.View()),
		m.messageInput.View(),
	) + "\n\n"
}

func createModel(handler *core.Handler, key any) ViewModel {
	textArea := textarea.New()
	textArea.Placeholder = "Compose a message"
	textArea.Focus()

	textArea.Prompt = "┃ "
	textArea.CharLimit = 280

	textArea.SetWidth(40)
	textArea.SetHeight(3)

	textArea.FocusedStyle.CursorLine = lipgloss.NewStyle()
	textArea.ShowLineNumbers = false

	viewport := viewport.New(40, 10)
	viewport.SetContent("You're online.")

	textArea.KeyMap.InsertNewline.SetEnabled(false)

	var aesGcmKey *crypto.AESGCMKey

	if newKey, ok := key.(*crypto.AESGCMKey); ok {
		aesGcmKey = newKey
	}

	return ViewModel{
		messageInput:     textArea,
		messages:         []UserMessage{},
		viewport:         viewport,
		currentUserStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		otherUserStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
		handler:          handler,
		aesGcmKey:        aesGcmKey,
		err:              nil,
	}
}

func viewportStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBottom(true).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		Width(width)
}
