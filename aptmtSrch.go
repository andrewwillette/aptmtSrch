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

func main() {
	go loadApartments()
	if err := tea.NewProgram(newModel()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func init() {
	initLog()
}

const listHeight = 9
const defaultWidth = 20

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
	choice        string
	quitting      bool
}

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	logg              *log.Logger
	apartmentsLoaded  = false
	loadedApartments  = []aptmtSrchr.Apartment{}
)

// loadApartments load apartments from internet and set local flags
// notifying runtime that apartments are loaded
func loadApartments() {
	defer func() { apartmentsLoaded = true }()
	loadedApartments = aptmtSrchr.GetApartments()
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
			l("pressed enter")
			i, ok := m.apartmentList.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				selectedUnit := getSelectedUnit(m.choice)
				l(fmt.Sprintf("selected unit: %+v", selectedUnit))
				for _, apt := range loadedApartments {
					if apt.UnitTitle == selectedUnit {
						openUrl(apt.ViewUrl)
					}
				}
			}
			return m, nil
		}
	default:
		if !apartmentsLoaded {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		} else {
			var cmd tea.Cmd
			m.apartmentList, cmd = m.apartmentList.Update(msg)
			return m, cmd
		}
	}
	if !apartmentsLoaded {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	} else {
		var cmd tea.Cmd
		m.apartmentList.SetItems(getApartmentUiItems(loadedApartments))
		m.apartmentList, cmd = m.apartmentList.Update(msg)
		return m, cmd
	}
}

var modelApartmentsSet = false

func l(tolog string) {
	logg.Println(tolog)
}

func (m model) View() string {

	// if m.choice != "" {
	// 	return "\n" + m.apartmentList.View()
	// }
	if m.quitting {
		return quitTextStyle.Render("Program Exited.")
	}
	if apartmentsLoaded {
		if !modelApartmentsSet {
			m.apartmentList.SetItems(getApartmentUiItems(loadedApartments))
			modelApartmentsSet = true
		}
		return fmt.Sprintf("\n%s\n\n", m.apartmentList.View())
	}

	return fmt.Sprintf("\n\n   %s Loading Apartments\n\n", m.spinner.View())
}

func openUrl(url string) {
	l(fmt.Sprintf("calling openUrl with url: %s", url))
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

func getEmptyApartmentUi() list.Model {
	items := []list.Item{}

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Apartments"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowPagination(true)
	l.SetShowHelp(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	return l
}

func getApartmentUiItems(aptmts []aptmtSrchr.Apartment) []list.Item {
	items := []list.Item{}
	for _, apt := range aptmts {
		items = append(items, item(fmt.Sprintf("%s : %s : %d", apt.AvailDate, apt.UnitTitle, apt.Rent)))
	}
	return items
}

func getSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}

func newModel() model {
	m := model{apartmentList: getEmptyApartmentUi(), spinner: getSpinner()}
	return m
}

func initLog() {
	var file, err1 = os.OpenFile("./bubTea.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err1 != nil {
		panic(err1)
	}
	logg = log.New(file, "", log.LstdFlags|log.Lshortfile)
}
