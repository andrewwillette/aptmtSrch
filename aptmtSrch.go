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

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	logg              *log.Logger
)

func init() {
	var file, err1 = os.OpenFile("./bubTea.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err1 != nil {
		panic(err1)
	}
	logg = log.New(file, "", log.LstdFlags|log.Lshortfile)
}

const listHeight = 16

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
	// spinner       spinner.Model
	choice   string
	quitting bool
}

func getApartments(apartmentListModel *[]item) {
}

func (m model) Init() tea.Cmd {
	// go func() { getApartments(&m.items) }()
	// return m.spinner.Tick
	return m.apartmentList.StartSpinner()
}

var aptmts []aptmtSrchr.Apartment

// func getApartments(aptmts *[]item

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	logg.Printf("calling Update\n")
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		logg.Println("m.Update tea.windowsizemsg case")
		m.apartmentList.SetWidth(msg.Width)
		logg.Println("m.Update tea.windowsizemsg case2")
		return m, nil

	case tea.KeyMsg:
		logg.Println("m.Update tea.keymsg case")
		switch keypress := msg.String(); keypress {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			// i, ok := m.apartment.SelectedItem().(item)
			// if ok {
			// 	m.choice = string(i)
			// 	selectedUnit := getSelectedUnit(m.choice)
			// 	for _, apt := range *aptmts {
			// 		if apt.UnitTitle == selectedUnit {
			// 			openUrl(apt.ViewUrl)
			// 		}
			// 	}
			// }
			return m, nil
		}
	}
	logg.Println("m.Update end of line")
	return m, nil
}

func (m model) View() string {
	logg.Printf("model.View()\n")
	// if m.choice != "" {
	// 	return "\n" + m.apartmentList.View()
	// }
	if m.quitting {
		return quitTextStyle.Render("Program Exited.")
	}

	// if len(*m.) >= 1 {
	// 	return fmt.Sprintf("\n%s\n\n", m.apartmentListModel.View())
	// 	// logg.Printf("apartments greater than 0\n")
	// 	// return "got apartments"
	// }
	return fmt.Sprintf("\n\n   %s Loading Apartments\n\n", m.apartmentList.View())
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

func getEmptyApartmentUi() list.Model {
	items := []list.Item{}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Available Apartments"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	return l
}

func getSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}

func newModel() model {
	m := model{apartmentList: getEmptyApartmentUi()}
	m.apartmentList.SetSpinner(getSpinner().Spinner)
	return m
}

func main() {
	if err := tea.NewProgram(newModel()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
