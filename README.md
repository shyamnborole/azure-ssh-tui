# azure-ssh-tui

A lightweight, strictly read-only terminal user interface for discovering Azure Virtual Machines and securely launching SSH sessions directly via the Azure CLI.

## Overview

`azure-ssh-tui` solves the problem of navigating the Azure Portal just to find a VM's IP address and SSH into it. It provides a beautiful, keyboard-driven interface to browse your Azure Subscriptions and VMs, fuzzy search for the VM you need, and SSH directly into it with a single keystroke.

Crucially, **this tool is 100% read-only**. It relies entirely on the local `az` CLI to read metadata and establish SSH connections. It contains zero code capable of modifying, deleting, or creating Azure resources. It does not implement its own authentication or manage SSH keys—it leverages your existing Azure CLI session.

## Features

- **Read-Only**: Safe for organizations to grant "Reader" access. Cannot mutate Azure resources.
- **No Custom Auth**: Leverages your existing `az login` context.
- **Fast Navigation**: Fuzzy search across VMs by name or Resource Group.
- **Subscription Switcher**: Easily switch between all your Azure subscriptions.
- **Direct SSH**: Press `s` or `Enter` to instantly `az ssh vm` into the selected machine.

## Requirements

1. **Azure CLI**: Must be installed and available in your `$PATH`.
2. **Authenticated Session**: You must run `az login` before using `azure-ssh-tui`.

## Installation

### From Source

```bash
git clone https://github.com/shyamborole/azure-ssh-tui.git
cd azure-ssh-tui
make build
./bin/azure-ssh-tui
```

### Pre-requisites

Ensure Azure CLI is installed:
```bash
az login
```

## Architecture

The application is built in Go using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.
It follows a clean architecture model:
- `cmd/azure-ssh-tui`: The application entry point.
- `internal/azure`: A strict wrapper around `az` CLI commands (`az account show`, `az vm list -d`, etc.).
- `internal/tui`: The terminal user interface logic, state machine, and components.

## Keyboard Shortcuts

| Key | Action |
| --- | --- |
| `↑` / `k` | Navigate Up |
| `↓` / `j` | Navigate Down |
| `Enter` | Launch SSH Session |
| `s` | Launch SSH Session |
| `/` | Search VMs |
| `Esc` | Clear Search / Close Modal |
| `Tab` | Switch Subscription |
| `r` | Refresh VMs |
| `q` | Quit |
| `?` | Help |

## Contributing

Contributions are welcome! Please ensure that any new features adhere strictly to the **read-only** philosophy. This application must never introduce write capabilities to Azure.
