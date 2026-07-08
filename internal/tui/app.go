package tui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shyamborole/azure-ssh-tui/internal/azure"
	"github.com/shyamborole/azure-ssh-tui/internal/config"
	"github.com/shyamborole/azure-ssh-tui/internal/tui/components"
)

type State int

const (
	StateLoading State = iota
	StateError
	StateVMList
	StateSubscriptionSwitch
	StateSSHConfirm
	StateConnectingAnimation
)

type App struct {
	client azure.Client

	state State

	width  int
	height int

	spinner    spinner.Model
	loadingMsg string
	err        error

	subs      []azure.Subscription
	activeSub *azure.Subscription

	vmTable components.VMTable
	subList components.SubList

	searchInput textinput.Model
	isSearching bool

	selectedVM   *azure.VM
	jumpHostName string
	jumpHostRG   string

	historyMode bool
	history     []config.HistoryEntry

	isSyncing   bool
	syncSpinner spinner.Model
	
	animFrame int
	targetCmd *exec.Cmd
}

func (a *App) SetHistoryMode(history []config.HistoryEntry) {
	a.historyMode = true
	a.history = history
}

func NewApp(client azure.Client) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 156
	ti.Width = 30

	cfg := config.Load()

	syncS := spinner.New()
	syncS.Spinner = spinner.MiniDot
	syncS.Style = SpinnerStyle

	return &App{
		client:       client,
		state:        StateLoading,
		loadingMsg:   "Connecting to Azure...",
		spinner:      s,
		syncSpinner:  syncS,
		vmTable:      components.NewVMTable(),
		subList:      components.NewSubList(),
		searchInput:  ti,
		jumpHostName: cfg.JumpHostName,
		jumpHostRG:   cfg.JumpHostResourceGroup,
	}
}

type authSuccessMsg struct{}
type errMsg struct{ err error }
type subsLoadedMsg struct{ subs []azure.Subscription }
type vmsLoadedMsg struct {
	vms   []azure.VM
	subID string
}
type animTickMsg struct{}
type sshFinishedMsg struct{ err error }

func (a *App) Init() tea.Cmd {
	if a.historyMode {
		return a.loadHistoryVMs()
	}

	cfg := config.Load()
	if len(cfg.CachedSubscriptions) > 0 {
		a.subs = cfg.CachedSubscriptions
		a.subList.SetSubscriptions(a.subs)

		for i, sub := range a.subs {
			if sub.IsDefault {
				a.activeSub = &a.subs[i]
				break
			}
		}
		if a.activeSub == nil && len(a.subs) > 0 {
			a.activeSub = &a.subs[0]
		}

		if a.activeSub != nil && cfg.CachedVMs != nil {
			if vms, ok := cfg.CachedVMs[a.activeSub.ID]; ok && len(vms) > 0 {
				a.vmTable.SetVMs(vms)
				a.state = StateVMList
				a.isSyncing = true

				return tea.Batch(
					a.syncSpinner.Tick,
					a.loadSubscriptions(),
					a.loadVMs(""),
				)
			}
		}
	}

	return tea.Batch(
		a.spinner.Tick,
		a.loadSubscriptions(),
		a.loadVMs(""),
	)
}

func (a *App) loadHistoryVMs() tea.Cmd {
	return func() tea.Msg {
		var vms []azure.VM
		for _, h := range a.history {
			vms = append(vms, azure.VM{
				Name:          h.Name,
				ResourceGroup: h.ResourceGroup,
			})
		}
		return vmsLoadedMsg{vms: vms, subID: ""}
	}
}

func (a *App) checkAuth() tea.Cmd {
	return func() tea.Msg {
		err := a.client.CheckAuth(context.Background())
		if err != nil {
			return errMsg{err: err}
		}
		return authSuccessMsg{}
	}
}

func (a *App) loadSubscriptions() tea.Cmd {
	return func() tea.Msg {
		subs, err := a.client.ListSubscriptions(context.Background())
		if err != nil {
			return errMsg{err: err}
		}
		return subsLoadedMsg{subs: subs}
	}
}

func (a *App) loadVMs(subID string) tea.Cmd {
	return func() tea.Msg {
		vms, err := a.client.ListVMs(context.Background(), subID)
		if err != nil {
			return errMsg{err: err}
		}
		return vmsLoadedMsg{vms: vms, subID: subID}
	}
}

func BuildSSHCmd(rg, vm string, preferPrivate bool, useJumpHost bool, jumpHostRG, jumpHostName string) *exec.Cmd {
	preferPrivateStr := ""
	if preferPrivate {
		preferPrivateStr = "--prefer-private-ip"
	}

	if useJumpHost && jumpHostName != "" {
		script := fmt.Sprintf(`tmpdir=$(mktemp -d) && az ssh config --resource-group '%s' --name '%s' --file "$tmpdir/config" >/dev/null 2>&1 && az ssh vm --resource-group '%s' --name '%s' %s -- -o StrictHostKeyChecking=no -o ProxyCommand="ssh -F '$tmpdir/config' -o StrictHostKeyChecking=no -W %%h:%%p '%s-%s' 2>/dev/null" 2>/dev/null ; rm -rf "$tmpdir"`,
			jumpHostRG, jumpHostName,
			rg, vm, preferPrivateStr,
			jumpHostRG, jumpHostName)

		return exec.Command("sh", "-c", script)
	}

	script := fmt.Sprintf(`az ssh vm --resource-group '%s' --name '%s' %s -- -o StrictHostKeyChecking=no 2>/dev/null`, rg, vm, preferPrivateStr)
	return exec.Command("sh", "-c", script)
}

func (a *App) launchSSH(rg, vm string, preferPrivate bool, useJumpHost bool) tea.Cmd {
	cfg := config.Load()
	cfg.JumpHostName = a.jumpHostName
	cfg.JumpHostResourceGroup = a.jumpHostRG
	cfg.LastVMName = vm
	cfg.LastVMResourceGroup = rg
	cfg.LastVMPreferPrivate = preferPrivate
	cfg.LastVMUsedJumpHost = useJumpHost

	cfg.AddHistory(config.HistoryEntry{
		Name:          vm,
		ResourceGroup: rg,
		PreferPrivate: preferPrivate,
		UsedJumpHost:  useJumpHost,
	})

	config.Save(cfg)

	a.targetCmd = BuildSSHCmd(rg, vm, preferPrivate, useJumpHost, a.jumpHostRG, a.jumpHostName)
	a.animFrame = 0
	a.state = StateConnectingAnimation

	return func() tea.Msg {
		return animTickMsg{}
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		listHeight := a.height - 4
		if listHeight < 0 {
			listHeight = 0
		}

		a.vmTable.SetHeight(listHeight)
		a.vmTable.SetWidth(a.width)

		a.subList.SetSize(a.width, a.height-2)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, Keys.ClearSearch):
			if a.isSearching {
				a.isSearching = false
				a.searchInput.Blur()
				a.searchInput.SetValue("")
				a.vmTable.ApplyFilter("")
				return a, nil
			} else if a.state == StateSubscriptionSwitch {
				a.state = StateVMList
				return a, nil
			} else if a.state == StateSSHConfirm {
				a.state = StateVMList
				a.selectedVM = nil
				return a, nil
			} else if a.state == StateError {
				return a, tea.Quit
			}
		}

	case authSuccessMsg:
		a.loadingMsg = "Loading subscriptions..."
		cmds = append(cmds, a.loadSubscriptions())

	case subsLoadedMsg:
		a.subs = msg.subs
		a.subList.SetSubscriptions(a.subs)

		// Find default or first
		for i, sub := range a.subs {
			if sub.IsDefault {
				a.activeSub = &a.subs[i]
				break
			}
		}
		if a.activeSub == nil && len(a.subs) > 0 {
			a.activeSub = &a.subs[0]
		}
		
		cfg := config.Load()
		cfg.CachedSubscriptions = a.subs
		config.Save(cfg)

		if a.activeSub == nil && len(a.subs) == 0 {
			a.err = fmt.Errorf("no subscriptions found")
			a.state = StateError
		}

	case vmsLoadedMsg:
		a.state = StateVMList
		a.vmTable.SetVMs(msg.vms)
		a.isSyncing = false
		
		subID := msg.subID
		if subID == "" && a.activeSub != nil {
			subID = a.activeSub.ID
		}
		if subID != "" {
			cfg := config.Load()
			if cfg.CachedVMs == nil {
				cfg.CachedVMs = make(map[string][]azure.VM)
			}
			cfg.CachedVMs[subID] = msg.vms
			config.Save(cfg)
		}
		
	case animTickMsg:
		a.animFrame++
		if a.animFrame >= 5 {
			cmd := tea.ExecProcess(a.targetCmd, func(err error) tea.Msg {
				return sshFinishedMsg{err: err}
			})
			cmds = append(cmds, cmd)
		} else {
			cmds = append(cmds, tea.Tick(time.Millisecond*400, func(t time.Time) tea.Msg {
				return animTickMsg{}
			}))
		}

	case errMsg:
		a.err = msg.err
		a.state = StateError

	case sshFinishedMsg:
		if msg.err != nil {
			a.err = fmt.Errorf("ssh session ended with error: %w", msg.err)
			a.state = StateError
		} else {
			// resume normal rendering, already in VMList probably
			a.state = StateVMList
		}
	}

	switch a.state {
	case StateLoading:
		a.spinner, cmd = a.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case StateVMList:
		if a.isSyncing {
			var spinCmd tea.Cmd
			a.syncSpinner, spinCmd = a.syncSpinner.Update(msg)
			cmds = append(cmds, spinCmd)
		}
		
		if a.isSearching {
			if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, Keys.Enter) {
				a.isSearching = false
				a.searchInput.Blur()
				return a, nil
			}
			a.searchInput, cmd = a.searchInput.Update(msg)
			cmds = append(cmds, cmd)
			a.vmTable.ApplyFilter(a.searchInput.Value())
		} else {
			switch msg := msg.(type) {
			case tea.KeyMsg:
				switch {
				case key.Matches(msg, Keys.Search):
					a.isSearching = true
					a.searchInput.Focus()
					return a, textinput.Blink
				case key.Matches(msg, Keys.SwitchSub):
					if !a.historyMode {
						a.state = StateSubscriptionSwitch
					}
				case key.Matches(msg, Keys.Refresh):
					if !a.historyMode && a.activeSub != nil {
						if !a.isSyncing {
							a.isSyncing = true
							cmds = append(cmds, a.syncSpinner.Tick, a.loadVMs(a.activeSub.ID))
						}
					}
				case key.Matches(msg, Keys.Enter), key.Matches(msg, Keys.SSH):
					vm := a.vmTable.SelectedVM()
					if vm != nil {
						a.selectedVM = vm
						a.state = StateSSHConfirm
					}
				case key.Matches(msg, Keys.SetJumpHost):
					vm := a.vmTable.SelectedVM()
					if vm != nil {
						a.jumpHostName = vm.Name
						a.jumpHostRG = vm.ResourceGroup
						cfg := config.Load()
						cfg.JumpHostName = a.jumpHostName
						cfg.JumpHostResourceGroup = a.jumpHostRG
						config.Save(cfg)
					}
				default:
					// Avoid updating table if clear search was pressed
					if !key.Matches(msg, Keys.ClearSearch) {
						var tableCmd tea.Cmd
						a.vmTable, tableCmd = a.vmTable.Update(msg)
						cmds = append(cmds, tableCmd)
					}
				}
			}
		}
	case StateSubscriptionSwitch:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if key.Matches(msg, Keys.Enter) {
				sub := a.subList.SelectedSubscription()
				if sub != nil {
					a.activeSub = sub
					a.state = StateVMList
					a.isSyncing = true
					
					cfg := config.Load()
					if vms, ok := cfg.CachedVMs[a.activeSub.ID]; ok && len(vms) > 0 {
						a.vmTable.SetVMs(vms)
					} else {
						// No cache for this sub, show empty or clear
						a.vmTable.SetVMs([]azure.VM{})
					}
					
					cmds = append(cmds, a.syncSpinner.Tick, a.loadVMs(a.activeSub.ID))
				}
			} else {
				var listCmd tea.Cmd
				a.subList, listCmd = a.subList.Update(msg)
				cmds = append(cmds, listCmd)
			}
		}
	case StateSSHConfirm:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "y" || msg.String() == "Y" {
				cmds = append(cmds, a.launchSSH(a.selectedVM.ResourceGroup, a.selectedVM.Name, true, false))
				a.state = StateConnectingAnimation
			} else if msg.String() == "n" || msg.String() == "N" {
				cmds = append(cmds, a.launchSSH(a.selectedVM.ResourceGroup, a.selectedVM.Name, false, false))
				a.state = StateConnectingAnimation
			} else if (msg.String() == "j" || msg.String() == "J") && a.jumpHostName != "" {
				cmds = append(cmds, a.launchSSH(a.selectedVM.ResourceGroup, a.selectedVM.Name, true, true))
				a.state = StateConnectingAnimation
			} else if key.Matches(msg, Keys.ClearSearch) {
				a.state = StateVMList
				a.selectedVM = nil
			}
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *App) View() string {
	if a.width == 0 {
		return "Initializing..."
	}

	switch a.state {
	case StateLoading:
		return fmt.Sprintf("\n\n   %s %s\n\n", a.spinner.View(), a.loadingMsg)
	case StateError:
		return fmt.Sprintf("\n\n%s\n\nPress 'q' to quit.", ErrorStyle.Render(a.err.Error()))
	case StateSubscriptionSwitch:
		return a.subList.View()
	case StateVMList:
		header := TitleStyle.Render("Azure SSH TUI")
		if a.historyMode {
			header += " | History"
		} else if a.activeSub != nil {
			header += fmt.Sprintf(" | Sub: %s", a.activeSub.Name)
		}
		if a.jumpHostName != "" {
			header += fmt.Sprintf(" | Jump: %s", a.jumpHostName)
		}

		if a.isSyncing {
			header += fmt.Sprintf("  [%s Syncing...]", a.syncSpinner.View())
		}

		var searchBar string
		if a.isSearching || a.searchInput.Value() != "" {
			searchBar = "\n" + a.searchInput.View()
		}

		table := a.vmTable.View()

		helpStr := "↑/↓: navigate • s/enter: ssh • J: set jump host • /: search • tab: switch sub • r: refresh • q: quit"
		help := renderFooter(helpStr, a.width)

		return fmt.Sprintf("%s%s\n%s\n%s", header, searchBar, table, help)
	case StateSSHConfirm:
		header := TitleStyle.Render("Azure SSH TUI")
		if a.activeSub != nil {
			header += fmt.Sprintf(" | Sub: %s", a.activeSub.Name)
		}
		if a.jumpHostName != "" {
			header += fmt.Sprintf(" | Jump: %s", a.jumpHostName)
		}

		promptStr := fmt.Sprintf("Connect to %s?\n\nPress 'y' for Yes (Private IP)\nPress 'n' for Yes (Public IP)", a.selectedVM.Name)
		if a.jumpHostName != "" {
			promptStr += fmt.Sprintf("\nPress 'j' for Yes (via Jump Host %s)", a.jumpHostName)
		}
		promptStr += "\nPress 'esc' to cancel"

		promptBox := BaseStyle.Copy().
			Padding(1, 2).
			BorderForeground(ColorAzureBlue).
			Render(promptStr)

		helpStr := "y: yes • n: no"
		if a.jumpHostName != "" {
			helpStr += " • j: jump host"
		}
		helpStr += " • esc: cancel"
		help := renderFooter(helpStr, a.width)

		return fmt.Sprintf("%s\n\n%s\n\n%s", header, promptBox, help)

	case StateConnectingAnimation:
		header := TitleStyle.Render("Azure SSH TUI")
		
		dots := strings.Repeat(".", (a.animFrame%4)+1)
		arrows := []string{"->", "-->", "--->", "---->", "----->"}
		arrow := arrows[a.animFrame%len(arrows)]
		
		var anim string
		if a.jumpHostName != "" {
			anim = fmt.Sprintf("Routing through Jump Host%s\n\n[Local] %s [Jump Host: %s] %s [Target: %s]", 
				dots, arrow, a.jumpHostName, arrow, a.selectedVM.Name)
		} else {
			anim = fmt.Sprintf("Connecting%s\n\n[Local] %s [Target: %s]", 
				dots, arrow, a.selectedVM.Name)
		}

		animBox := BaseStyle.Copy().
			Padding(2, 4).
			BorderForeground(ColorSuccess).
			Render(anim)

		return fmt.Sprintf("%s\n\n%s", header, animBox)
	}

	return ""
}

func renderFooter(left string, width int) string {
	right := "By SHYAM"
	pad := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if pad < 0 {
		pad = 0
	}
	return StatusBarStyle.Width(width).Render(left + strings.Repeat(" ", pad) + right)
}
