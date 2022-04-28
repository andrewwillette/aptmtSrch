package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/andrewwillette/aptmtSrchr"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 16

var aptmts = aptmtSrchr.GetApartments()

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprintf(w, fn(str))
}

type model struct {
	apartmentList list.Model
	spinner       spinner.Model
	items         []item
	choice        string
	quitting      bool
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.apartmentList.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.apartmentList.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				selectedUnit := getSelectedUnit(m.choice)
				for _, apt := range aptmts {
					if apt.UnitTitle == selectedUnit {
						openUrl(apt.ViewUrl)
					}
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.apartmentList, cmd = m.apartmentList.Update(msg)
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m model) View() string {
	// if m.choice != "" {
	// 	return "\n" + m.apartmentList.View()
	// }
	if m.quitting {
		return quitTextStyle.Render("Program Exited.")
	}
	return fmt.Sprintf("\n%s\n\n%s", m.apartmentList.View(), m.spinner.View())
}

func openUrl(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
func getSelectedUnit(selected string) string {
	r, _ := regexp.Compile(`\s[^:].*:`)
	result := r.FindString(selected)
	r, _ = regexp.Compile(`.*[^:]`)
	result = r.FindString(result)
	return strings.TrimSpace(result)
}
func main() {
	items := []list.Item{}
	for _, apt := range aptmts {
		items = append(items, item(fmt.Sprintf("%s : %s : %d", apt.AvailDate, apt.UnitTitle, apt.Rent)))
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Available Apartments"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{apartmentList: l}

	if err := tea.NewProgram(m).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
