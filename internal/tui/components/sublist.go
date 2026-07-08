package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shyamborole/azure-ssh-tui/internal/azure"
)

type SubItem struct {
	Sub azure.Subscription
}

func (i SubItem) Title() string       { return i.Sub.Name }
func (i SubItem) Description() string { return i.Sub.ID }
func (i SubItem) FilterValue() string { return i.Sub.Name }

type SubList struct {
	List list.Model
}

func NewSubList() SubList {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		BorderLeftForeground(lipgloss.Color("229"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		BorderLeftForeground(lipgloss.Color("229"))

	l := list.New(nil, delegate, 0, 0)
	l.Title = "Select Subscription"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = l.Styles.Title.Background(lipgloss.Color("#0078D4"))

	return SubList{
		List: l,
	}
}

func (m *SubList) Init() tea.Cmd {
	return nil
}

func (m *SubList) Update(msg tea.Msg) (SubList, tea.Cmd) {
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return *m, cmd
}

func (m *SubList) View() string {
	return m.List.View()
}

func (m *SubList) SetSize(width, height int) {
	m.List.SetSize(width, height)
}

func (m *SubList) SetSubscriptions(subs []azure.Subscription) {
	items := make([]list.Item, len(subs))
	for i, sub := range subs {
		items[i] = SubItem{Sub: sub}
	}
	m.List.SetItems(items)
}

func (m *SubList) SelectedSubscription() *azure.Subscription {
	i, ok := m.List.SelectedItem().(SubItem)
	if ok {
		return &i.Sub
	}
	return nil
}
