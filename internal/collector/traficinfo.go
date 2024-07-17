package collector

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type connections struct {
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

func getConnectionInfo(protocol, file string) ([]connections, error) {

	var objectConnection []connections

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

		objectConnection = append(objectConnection, connections{
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

func aggregateInfo(M time.Duration) ([]connections, map[string]int, map[string]int) {
	ticker := time.NewTicker(M)
	defer ticker.Stop()
	var aggregateSlice []connections
	statisticsMap := make(map[string]*connections)
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

func TrafficGetInfo() ([]NetworkProtocol, []connections, []TCPStates) {
	var totalBytes int
	var networkProtocol []NetworkProtocol
	var tcpState []TCPStates
	connects, protoStat, statTCP := aggregateInfo(time.Second)

	for _, tbytes := range protoStat {
		totalBytes += tbytes
	}
	for protocol, bytes := range protoStat {
		percent := bytes / totalBytes * 100
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
	return networkProtocol, connects, tcpState
}
