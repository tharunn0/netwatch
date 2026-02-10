package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xruc/netwatch/conn"
)

// Update handles messages and updates the model state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r":
			return m, fetchConnections(m.netPath) // manual refresh
		case "l":
			// Toggle filter mode: all -> local -> public -> all
			switch m.filterMode {
			case FilterAll:
				m.filterMode = FilterLocal
			case FilterLocal:
				m.filterMode = FilterPublic
			case FilterPublic:
				m.filterMode = FilterAll
			}
			// Reset cursor when filter changes
			m.cursor = 0
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			// We'll check bounds in the view based on filtered connections
			if m.cursor < len(m.connections)-1 {
				m.cursor++
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case []conn.Connection:
		// Update connections when new data arrives
		m.connections = msg
		m.err = nil
		// Keep cursor in bounds
		if m.cursor >= len(msg) {
			m.cursor = len(msg) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}

	case errMsg:
		m.err = msg.err

	case tickMsg:
		// Periodic refresh
		return m, tea.Batch(
			fetchConnections(m.netPath),
			tickEvery(),
		)
	}

	m.connections = m.filterConnections()

	return m, nil
}

// filterConnections returns connections based on the current filter mode
func (m Model) filterConnections() []conn.Connection {
	if m.filterMode == FilterAll {
		return m.connections
	}

	filtered := make([]conn.Connection, 0)
	for _, c := range m.connections {
		// Only check the remote address to determine if connection is local or public
		isLocal := isLocalAddress(c.RemoteIp)

		if m.filterMode == FilterLocal && isLocal {
			filtered = append(filtered, c)
		} else if m.filterMode == FilterPublic && !isLocal {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

// isLocalAddress checks if an IP address is local/loopback/private
func isLocalAddress(ip string) bool {
	// Loopback and unspecified addresses
	if ip == "127.0.0.1" || ip == "::1" || ip == "0.0.0.0" {
		return true
	}

	// Loopback range (127.x.x.x)
	if strings.HasPrefix(ip, "127.") {
		return true
	}

	// Private IP ranges
	if strings.HasPrefix(ip, "10.") {
		return true
	}

	if strings.HasPrefix(ip, "192.168.") {
		return true
	}

	// 172.16.0.0 - 172.31.255.255
	if strings.HasPrefix(ip, "172.") {
		parts := strings.Split(ip, ".")
		if len(parts) >= 2 {
			second := parts[1]
			// Check if second octet is between 16 and 31
			for i := 16; i <= 31; i++ {
				if second == fmt.Sprintf("%d", i) {
					return true
				}
			}
		}
	}

	// Link-local (169.254.x.x)
	if strings.HasPrefix(ip, "169.254.") {
		return true
	}

	// Multicast (224.x.x.x - 239.x.x.x)
	if strings.HasPrefix(ip, "224.") || strings.HasPrefix(ip, "225.") ||
		strings.HasPrefix(ip, "226.") || strings.HasPrefix(ip, "227.") ||
		strings.HasPrefix(ip, "228.") || strings.HasPrefix(ip, "229.") ||
		strings.HasPrefix(ip, "230.") || strings.HasPrefix(ip, "231.") ||
		strings.HasPrefix(ip, "232.") || strings.HasPrefix(ip, "233.") ||
		strings.HasPrefix(ip, "234.") || strings.HasPrefix(ip, "235.") ||
		strings.HasPrefix(ip, "236.") || strings.HasPrefix(ip, "237.") ||
		strings.HasPrefix(ip, "238.") || strings.HasPrefix(ip, "239.") {
		return true
	}

	return false
}
