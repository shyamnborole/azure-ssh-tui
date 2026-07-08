package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shyamborole/azure-ssh-tui/internal/azure"
)

type VMTable struct {
	table    table.Model
	vms      []azure.VM
	filtered []azure.VM
	filter   string
}

func NewVMTable() VMTable {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Resource Group", Width: 20},
		{Title: "Status", Width: 15},
		{Title: "Public IP", Width: 15},
		{Title: "Private IP", Width: 15},
		{Title: "Location", Width: 15},
		{Title: "OS", Width: 10},
		{Title: "Size", Width: 15},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return VMTable{
		table: t,
	}
}

func (m *VMTable) Init() tea.Cmd {
	return nil
}

func (m *VMTable) Update(msg tea.Msg) (VMTable, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return *m, cmd
}

func (m *VMTable) View() string {
	return m.table.View()
}

func (m *VMTable) SetVMs(vms []azure.VM) {
	// Preserve currently selected VM if possible
	var selectedName string
	if sel := m.SelectedVM(); sel != nil {
		selectedName = sel.Name
	}
	
	m.vms = vms
	m.ApplyFilter(m.filter)
	
	if selectedName != "" {
		for i, vm := range m.filtered {
			if vm.Name == selectedName {
				m.table.SetCursor(i)
				break
			}
		}
	}
}

func (m *VMTable) ApplyFilter(filter string) {
	m.filter = strings.ToLower(filter)
	m.filtered = make([]azure.VM, 0)
	
	for _, vm := range m.vms {
		if filter == "" || 
			strings.Contains(strings.ToLower(vm.Name), m.filter) ||
			strings.Contains(strings.ToLower(vm.ResourceGroup), m.filter) {
			m.filtered = append(m.filtered, vm)
		}
	}
	
	m.updateRows()
}

func (m *VMTable) updateRows() {
	var rows []table.Row
	for _, vm := range m.filtered {
		status := vm.PowerState
		if status == "" {
			status = "Unknown"
		} else {
			status = strings.TrimPrefix(status, "VM ")
		}

		size := vm.HardwareProfile.VMSize
		
		rows = append(rows, table.Row{
			vm.Name,
			vm.ResourceGroup,
			status,
			vm.PublicIps,
			vm.PrivateIps,
			vm.Location,
			vm.OS(),
			size,
		})
	}
	m.table.SetRows(rows)
}

func (m *VMTable) SelectedVM() *azure.VM {
	if len(m.filtered) == 0 {
		return nil
	}
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.filtered) {
		return &m.filtered[cursor]
	}
	return nil
}

func (m *VMTable) SetHeight(h int) {
	m.table.SetHeight(h)
}

func (m *VMTable) SetWidth(w int) {
	// Table doesn't natively support SetWidth dynamically to change column widths,
	// but we can at least call SetWidth to ensure it fits the viewport if needed.
	// We might need to recalculate column widths manually if we want responsive columns.
}
