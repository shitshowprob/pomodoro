package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	TotalDuration      time.Duration
	DurationLeft       time.Duration
	TotalBreakDuration time.Duration
	sessionCount       int
	display            string
}

func NewModel(duration time.Duration) *model {
	return &model{
		TotalDuration:      duration,
		TotalBreakDuration: time.Minute * 5,
		DurationLeft:       duration,
	}
}

type TickMsg time.Time

var timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "s":
			m.display = fmt.Sprintf("Starting a pomodoro session of %s and %s break\n", m.TotalDuration.String(), m.TotalBreakDuration.String())
			return m, m.Tick()
		}
	case TickMsg:
		m.display = fmt.Sprintf("%s", m.DurationLeft.String())
		return m, m.Tick()
	}
	return m, nil
}

func (m *model) View() string {
	return timeStyle.Render(m.display)
}

func (m *model) Tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		m.DurationLeft = m.DurationLeft - time.Second
		// m.DurationLeft = m.DurationLeft.Abs()
		return TickMsg(t)
	})
}

func main() {
	m := NewModel(time.Minute * 25)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
