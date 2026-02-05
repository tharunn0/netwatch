package conn

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Connection struct {
	LocalIp    string
	LocalPort  string
	RemoteIp   string
	RemotePort string
	State      string
	Inode      string

	Proc string
	PID  string
}

func FetchConnections(netPath string) ([]Connection, error) {

	// fetching all the live tcp connections
	file, err := os.Open(netPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var conns []Connection

	scanner.Scan()

	for scanner.Scan() {

		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}

		localIp, localPort := parse(fields[1])
		remoteIp, remotePort := parse(fields[2])

		state := tcpState(fields[3])
		inode := fields[9]

		conns = append(conns, Connection{
			LocalIp:    localIp,
			LocalPort:  localPort,
			RemoteIp:   remoteIp,
			RemotePort: remotePort,
			State:      state,
			Inode:      inode,
		})

	}

	// fetching and mapping processes to inodes
	uniqueInodes := make(map[string]struct{}, len(conns))

	for _, c := range conns {
		uniqueInodes[c.Inode] = struct{}{}
	}

	inodeToProc := make(map[string]struct {
		pid  string
		proc string
	})

	// read all the entries inside proc
	procEntries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	permissionDenied := false

	for _, p := range procEntries {

		// exit if all inodes are resolved
		if len(uniqueInodes) == 0 {
			break
		}

		pid := p.Name()

		// filter out the processes that does not start with a integer
		if pid[0] < '0' || pid[0] > '9' {
			continue
		}

		fdDir := filepath.Join("/proc", pid, "fd")

		fdEntries, err := os.ReadDir(fdDir)
		if err != nil {
			// check if its permission issues
			if os.IsPermission(err) {
				permissionDenied = true
			}
			continue
		}

		commBytes, err := os.ReadFile(filepath.Join("/proc", pid, "comm"))
		if err != nil {
			continue
		}

		procName := strings.TrimSpace(string(commBytes))

		for _, fd := range fdEntries {

			if len(uniqueInodes) == 0 {
				break
			}
			fdPath := filepath.Join(fdDir, fd.Name())

			link, err := os.Readlink(fdPath)
			if err != nil {
				continue
			}

			if !strings.HasPrefix(link, "socket:[") {
				continue
			}

			inode := strings.TrimSuffix(
				strings.TrimPrefix(link, "socket:["),
				"]",
			)

			if _, ok := uniqueInodes[inode]; !ok {
				continue
			}

			inodeToProc[inode] = struct {
				pid  string
				proc string
			}{
				pid:  pid,
				proc: procName,
			}

			delete(uniqueInodes, inode)
		}

	}

	for i := range conns {

		if conns[i].PID == "0" {
			conns[i].Proc = "No Owner"
			continue
		}

		if p, ok := inodeToProc[conns[i].Inode]; ok {
			conns[i].PID = p.pid
			conns[i].Proc = p.proc
			continue
		}

		if permissionDenied {
			conns[i].Proc = "Permission Denied"
		} else {
			conns[i].Proc = "Unknown"
		}

	}

	return conns, nil
}

// sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
// "0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345"

func parse(hex string) (string, string) {
	parts := strings.Split(hex, ":")
	iphex, porthex := parts[0], parts[1]

	ip := getip(iphex)

	port, _ := strconv.ParseInt(porthex, 16, 16)

	return ip, fmt.Sprintf("%d", port)
}

func getip(ip string) string {
	h, _ := hex.DecodeString(ip)
	for i, j := 0, len(h)-1; i < j; i, j = i+1, j-1 {
		h[i], h[j] = h[j], h[i]
	}

	return net.IP(h).String()
}

func tcpState(code string) string {
	states := map[string]string{
		"01": "ESTABLISHED",
		"02": "SYN_SENT",
		"03": "SYN_RECV",
		"04": "FIN_WAIT1",
		"05": "FIN_WAIT2",
		"06": "TIME_WAIT",
		"07": "CLOSE",
		"08": "CLOSE_WAIT",
		"09": "LAST_ACK",
		"0A": "LISTEN",
		"0B": "CLOSING",
	}

	if s, ok := states[code]; ok {
		return s
	}
	return code
}
