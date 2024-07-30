package collector

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TrafficInfo struct {
	SourceIP   string
	SourcePort int
	DestIP     string
	DestPort   int
	Protocol   string
	Bytes      int
	State      string
	BPS        float64
}

func parsHex(hex string) (string, int) {
	hexIP := hex[:8]
	hexPort := hex[8:]
	ip := ""
	for i := 0; i < 8; i += 2 {
		octd, err := strconv.ParseInt(hexIP[i:i+2], 16, 32)
		if err != nil {
			return "", 0
		}
		ip = fmt.Sprintf("%s%d.", ip, octd)
	}
	ip = strings.TrimRight(ip, ".")
	port, err := strconv.ParseInt(hexPort, 16, 32)
	if err != nil {
		return "", 0
	}
	return ip, int(port)
}

func getConnectionInfo(protocol, file string) ([]TrafficInfo, error) {

	var objectConnection []TrafficInfo

	getInfo, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer getInfo.Close()
	scanner := bufio.NewScanner(getInfo)
	scanner.Scan()
	for scanner.Scan() {
		var (
			sl, localAddr, remAddr, state string
			txQueue, rxQueue, timer       int
			retrnsmt, uid, timeout        int
			inode                         uint64
		)
		fmt.Sscan(scanner.Text(), &sl, &localAddr, &remAddr, &state, &txQueue, &rxQueue, &timer, &retrnsmt, &uid, &timeout, &inode)
		LAddr, LPort := parsHex(localAddr)
		RAddr, RPort := parsHex(remAddr)

		objectConnection = append(objectConnection, TrafficInfo{
			SourceIP:   LAddr,
			SourcePort: LPort,
			DestIP:     RAddr,
			DestPort:   RPort,
			Protocol:   protocol,
			Bytes:      rxQueue,
			State:      state,
		})
	}
	return objectConnection, nil
}

func aggregateInfo(M time.Duration) ([]TrafficInfo, map[string]int, map[string]int) {
	ticker := time.NewTicker(M)
	defer ticker.Stop()
	var aggregateSlice []TrafficInfo
	statisticsMap := make(map[string]*TrafficInfo)
	protocolBytesMap := make(map[string]int)
	protocolStateMap := make(map[string]int)
	protocolMap := map[string]string{
		"tcp":  "/proc/net/tcp",
		"udp":  "/proc/net/udp",
		"icmp": "/proc/net/icmp",
	}
	run := time.Now()
	for range ticker.C {
		for protocol, file := range protocolMap {
			getConn, err := getConnectionInfo(protocol, file)
			if err != nil {
				fmt.Printf("Error getting %s connections: %v\n", protocol, err)
				continue
			}
			for _, conn := range getConn {
				key := fmt.Sprintf("%s:%d-%s:%d-%s", conn.SourceIP, conn.SourcePort, conn.DestIP, conn.DestPort, conn.Protocol)
				if existConn, ok := statisticsMap[key]; ok {
					existConn.Bytes += conn.Bytes
					statisticsMap[key] = existConn
				} else {
					statisticsMap[key] = &conn
				}
				protocolBytesMap[protocol] += conn.Bytes
				if protocol == "tcp" {
					protocolStateMap[conn.State]++
				}
			}
		}
		if time.Since(run) >= M {
			break
		}
	}
	for _, conn := range statisticsMap {
		conn.BPS = float64(conn.Bytes) / float64(time.Second)
		aggregateSlice = append(aggregateSlice, *conn)
	}
	sort.Slice(aggregateSlice, func(i, j int) bool {
		return aggregateSlice[i].BPS > aggregateSlice[j].BPS
	})
	return aggregateSlice, protocolBytesMap, protocolStateMap
}

func getListeningSockets() ([]ListeningSocket, error) {
	var listeningSockets []ListeningSocket

	files := []string{"/proc/net/tcp", "/proc/net/tcp6"}
	for _, file := range files {
		sockets, err := parseSockets(file)
		if err != nil {
			return nil, err
		}

		for _, socket := range sockets {
			protocol := "tcp"
			if file == "/proc/net/tcp6" {
				protocol = "tcp6"
			}

			user, err := getUserFromPID(socket.PID)
			if err != nil {
				user = "unknown"
			}

			ls := ListeningSocket{
				Command:  socket.Command,
				PID:      socket.PID,
				User:     user,
				Protocol: protocol,
				Port:     socket.Port,
			}

			listeningSockets = append(listeningSockets, ls)
		}
	}

	return listeningSockets, nil
}

func getUserFromPID(pid int) (string, error) {
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdlineFile, err := os.Open(cmdlinePath)
	if err != nil {
		return "", err
	}
	defer cmdlineFile.Close()

	cmdlineScanner := bufio.NewScanner(cmdlineFile)
	cmdlineScanner.Scan()
	// cmdline := cmdlineScanner.Text()
	// command := strings.Split(cmdline, "\x00")[0]

	uidPath := fmt.Sprintf("/proc/%d/status", pid)
	statusFile, err := os.Open(uidPath)
	if err != nil {
		return "unknown", err
	}
	defer statusFile.Close()

	var uidString string
	statusScanner := bufio.NewScanner(statusFile)
	for statusScanner.Scan() {
		text := statusScanner.Text()
		if strings.HasPrefix(text, "Uid:") {
			uidString = strings.Fields(text)[1]
			break
		}
	}

	uid, err := strconv.Atoi(uidString)
	if err != nil {
		return "unknown", err
	}

	u, err := user.LookupId(fmt.Sprintf("%d", uid))
	if err != nil {
		return "unknown", err
	}

	return u.Username, nil
}

type socketInfo struct {
	Command string
	PID     int
	Port    int
}

func parseSockets(file string) ([]socketInfo, error) {
	var sockets []socketInfo

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		localAddress := fields[1]
		parts := strings.Split(localAddress, ":")
		portHex := parts[1]
		port, err := strconv.ParseInt(portHex, 16, 32)
		if err != nil {
			return nil, err
		}
		inodeStr := fields[9]

		pid, command, err := findProcessByInode(inodeStr)
		if err != nil {
			pid, command = -1, "unknown"
		}

		socket := socketInfo{
			Command: command,
			PID:     pid,
			Port:    int(port),
		}
		sockets = append(sockets, socket)
	}

	return sockets, nil
}

func findProcessByInode(inode string) (int, string, error) {
	fdDir := "/proc"
	d, err := os.Open(fdDir)
	if err != nil {
		return -1, "", err
	}
	defer d.Close()

	processDirs, err := d.Readdirnames(-1)
	if err != nil {
		return -1, "", err
	}

	for _, pidDir := range processDirs {
		if _, err := strconv.Atoi(pidDir); err == nil {
			fdPath := fmt.Sprintf("%s/%s/fd", fdDir, pidDir)
			fds, err := os.ReadDir(fdPath)
			if err != nil {
				continue
			}

			for _, fd := range fds {
				link, err := os.Readlink(filepath.Join(fdPath, fd.Name()))
				if err != nil {
					continue
				}
				if strings.Contains(link, inode) {
					cmd, err := os.ReadFile(fmt.Sprintf("/proc/%s/comm", pidDir))
					if err != nil {
						return -1, "", err
					}
					command := strings.TrimSpace(string(cmd))
					pid, err := strconv.Atoi(pidDir)
					if err != nil {
						return -1, "", err
					}
					return pid, command, nil
				}
			}
		}
	}
	return -1, "", fmt.Errorf("inode not found")
}

func TrafficGetInfo() ([]NetworkProtocol, []TrafficInfo, []TCPStates, []ListeningSocket) {
	var totalBytes int
	var percent int
	var networkProtocol []NetworkProtocol
	var tcpState []TCPStates

	connects, protoStat, statTCP := aggregateInfo(time.Second)
	for _, tbytes := range protoStat {
		totalBytes += tbytes
	}
	for protocol, bytes := range protoStat {
		if totalBytes != 0 {
			percent = bytes / totalBytes * 100
		} else {
			percent = bytes
		}

		networkProtocol = append(networkProtocol, NetworkProtocol{
			Protocol: protocol,
			Bytes:    int64(bytes),
			Percent:  float64(percent),
		})
	}
	for state, count := range statTCP {
		tcpState = append(tcpState, TCPStates{
			State: state,
			Count: count,
		})
	}

	listeningSockets, err := getListeningSockets()
	if err != nil {
		fmt.Printf("Error getting listening sockets: %v\n", err)
	}

	return networkProtocol, connects, tcpState, listeningSockets
}
