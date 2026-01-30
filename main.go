package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).PaddingTop(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	timeStyle         = lipgloss.NewStyle().PaddingLeft(4).PaddingTop(2).Foreground(lipgloss.Color("205"))
)

type rest time.Duration
type session time.Duration

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := string(i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	List          list.Model
	DurationLeft  time.Duration
	BreakDuration time.Duration
	listToggle    bool
	breakStart    bool
	paused        bool
	sessionCount  int
	display       string
}

type TickMsg struct{}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.List.SetWidth(msg.Width)
		return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "q", tea.KeyCtrlC.String():
			cmd = tea.Printf("Completed sessions: %d ", m.sessionCount)
			return m, tea.Batch(tea.Quit, cmd)
		case "e", tea.KeyEnter.String():
			it := m.List.SelectedItem()
			i, ok := it.(item)
			if !ok {
				cmd = tea.Println("A failure Occured, quitting...")
				return m, tea.Batch(cmd)
			}
			d := strings.SplitN(string(i), "/", 2)
			minutes, _ := strconv.Atoi(d[0])
			breakMinutes, _ := strconv.Atoi(d[1])
			m.Reset()
			m.listToggle = false
			m.DurationLeft = time.Minute * time.Duration(minutes)
			m.BreakDuration = time.Minute * time.Duration(breakMinutes)
			m.display = fmt.Sprintf("Starting a pomodoro session of %s and %s break\n", m.DurationLeft.String(), m.BreakDuration.String())
			return m, m.Tick()
		case tea.KeyEsc.String(), "p":
			if m.paused {
				m.paused = false
				m.listToggle = false
				return m, m.Tick()
			} else {
				m.paused = true
				m.listToggle = true
			}
		}
	case TickMsg:
		if m.paused {
			return m, cmd
		}
		if m.DurationLeft > 5*time.Second {
			m.display = fmt.Sprintf("🍅 %s", m.DurationLeft.String())
		} else {
			if m.DurationLeft == 0 && m.breakStart {
				m.breakStart = false
				m.sessionCount++
				m.display = ""
				m.List, cmd = m.List.Update(msg)
				return m, cmd
			} else if m.DurationLeft == 0 && !m.breakStart {
				m.breakStart = true
				m.DurationLeft = m.BreakDuration
				m.display = fmt.Sprintf("No of Sessions Completed:%d", m.sessionCount)
			}
			add := "Session"
			if m.breakStart {
				add = "😴Break"
			}
			m.display = fmt.Sprintf("%s ending in....%s", add, m.DurationLeft.String())
		}
		return m, m.Tick()
	}

	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	var sb strings.Builder
	if m.listToggle {
		sb.WriteString(m.List.View())
		sb.WriteString("\n")
	}
	if m.paused {
		sb.WriteString(itemStyle.Render("Paused the timer: "))
	}
	sb.WriteString(timeStyle.Render(m.display))
	return sb.String()
}

func (m *model) Tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		m.DurationLeft = m.DurationLeft - time.Second
		return TickMsg{}
	})
}

func (m *model) Reset() {
	m = NewModel()
}

func NewModel() *model {
	items := []list.Item{
		item("25/5"), item("50/10"),
	}
	li := list.New(items, itemDelegate{}, 20, len(items)*4)
	li.SetFilteringEnabled(false)
	li.SetShowStatusBar(false)
	li.SetShowFilter(false)
	li.SetShowHelp(false)
	li.Title = "Choose the pomodoro setting"
	li.Styles.Title = titleStyle
	return &model{
		List:          li,
		BreakDuration: time.Minute * 5,
		DurationLeft:  time.Minute * 25,
		listToggle:    true,
	}
}
func main() {
	m := NewModel()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
