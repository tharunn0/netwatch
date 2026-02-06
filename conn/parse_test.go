package conn

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input        string
		expectedIP   string
		expectedPort string
	}{
		{"0100007F:1F90", "127.0.0.1", "8080"},
		{"00000000:0000", "0.0.0.0", "0"},
		{"01020304:0050", "4.3.2.1", "80"},
	}

	for _, tt := range tests {
		ip, port := parse(tt.input)
		if ip != tt.expectedIP {
			t.Errorf("parse(%q) IP = %v, want %v", tt.input, ip, tt.expectedIP)
		}
		if port != tt.expectedPort {
			t.Errorf("parse(%q) Port = %v, want %v", tt.input, port, tt.expectedPort)
		}
	}
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0100007F", "127.0.0.1"},
		{"00000000", "0.0.0.0"},
		{"01020304", "4.3.2.1"},
		{"FFFFFFFF", "255.255.255.255"},
	}

	for _, tt := range tests {
		result := getip(tt.input)
		if result != tt.expected {
			t.Errorf("getip(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestTcpState(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"01", "ESTABLISHED"},
		{"02", "SYN_SENT"},
		{"03", "SYN_RECV"},
		{"04", "FIN_WAIT1"},
		{"05", "FIN_WAIT2"},
		{"06", "TIME_WAIT"},
		{"07", "CLOSE"},
		{"08", "CLOSE_WAIT"},
		{"09", "LAST_ACK"},
		{"0A", "LISTEN"},
		{"0B", "CLOSING"},
		{"99", "99"}, // Unknown state
	}

	for _, tt := range tests {
		result := tcpState(tt.input)
		if result != tt.expected {
			t.Errorf("tcpState(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
