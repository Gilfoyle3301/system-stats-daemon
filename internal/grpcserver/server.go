package grpcserver

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	collectorpb "github.com/Gilfoyle3301/system-stats-daemon/api/pb"
	"github.com/Gilfoyle3301/system-stats-daemon/internal/collector"
	"google.golang.org/grpc"
)

type MetricsCollectorServer struct {
	collectorpb.UnimplementedMetricsCollectorServer
	collector *collector.Collector
}

func (s *MetricsCollectorServer) CollectMetrics(req *collectorpb.MetricsRequest, stream collectorpb.MetricsCollector_CollectMetricsServer) error {
	period := time.Duration(req.GetNSecond()) * time.Second
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		dataList []*collector.Collector
		mu       sync.Mutex
	)
	go func() {
		for {
			select {
			case <-ticker.C:
				data := collector.Collect()
				mu.Lock()
				dataList = append(dataList, data)
				mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			mu.Lock()
			averageData := computeAverages(dataList)
			mu.Unlock()

			response := &collectorpb.MetricsResponse{
				Collector: averageData,
			}

			if err := stream.Send(response); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

func computeAverages(dataList []*collector.Collector) *collectorpb.Collector {
	count := len(dataList)
	if count == 0 {
		return &collectorpb.Collector{}
	}
	avgLoad := &collectorpb.LoadAverage{}
	avgCPU := &collectorpb.CPUUsage{}
	avgDisk := []*collectorpb.DiskUsage{}
	avgConnections := []*collectorpb.TrafficInfo{}
	avgTCPState := []*collectorpb.TCPStates{}
	avgListeningSockets := []*collectorpb.ListeningSocket{}
	avgFileSystemUsages := []*collectorpb.FileSystemUsage{}
	protocolBytes := make(map[string]int64)
	for _, metrics := range dataList {
		avgLoad.OneMinute += metrics.LoadAverage.OneMinute
		avgLoad.FiveMinutes += metrics.LoadAverage.FiveMinutes
		avgLoad.FifteenMinutes += metrics.LoadAverage.FifteenMinutes

		avgCPU.UserMode += metrics.CPUUsage.UserMode
		avgCPU.SystemMode += metrics.CPUUsage.SystemMode
		avgCPU.Idle += metrics.CPUUsage.Idle

		avgLoad.OneMinute /= float64(count)
		avgLoad.FiveMinutes /= float64(count)
		avgLoad.FifteenMinutes /= float64(count)

		avgCPU.UserMode /= float64(count)
		avgCPU.SystemMode /= float64(count)
		avgCPU.Idle /= float64(count)

		for _, diskState := range metrics.DiskUsage {
			found := false
			for _, avgState := range avgDisk {
				if avgState.Name == diskState.Name {
					avgState.Kbpersec += diskState.KBPerSec
					avgState.Tps += diskState.TPS
					found = true
					break
				}
			}
			if !found {
				avgDisk = append(avgDisk, &collectorpb.DiskUsage{
					Name:     diskState.Name,
					Tps:      diskState.TPS,
					Kbpersec: diskState.KBPerSec,
				})
			}
		}

		for _, conn := range metrics.TrafficInfo {
			found := false
			for _, avgConn := range avgConnections {
				if avgConn.Sourceip == conn.SourceIP && int(avgConn.SourcePort) == int(conn.SourcePort) &&
					avgConn.Destip == conn.DestIP && int(avgConn.DestPort) == int(conn.DestPort) &&
					avgConn.Protocol == conn.Protocol && avgConn.State == conn.State {
					avgConn.Bytes += int64(conn.Bytes)
					found = true
					break
				}
			}
			if !found {
				avgConnections = append(avgConnections, &collectorpb.TrafficInfo{
					Sourceip:   conn.SourceIP,
					SourcePort: int64(conn.SourcePort),
					Destip:     conn.DestIP,
					DestPort:   int64(conn.DestPort),
					Protocol:   conn.Protocol,
					Bps:        conn.BPS,
					Bytes:      int64(conn.Bytes),
					State:      conn.State,
				})
			}
		}

		for _, socket := range metrics.ListeningSocket {
			found := false
			for _, avgSocket := range avgListeningSockets {
				if avgSocket.Pid == int64(socket.PID) && avgSocket.Command == socket.Command && avgSocket.User == socket.User {
					found = true
					break
				}
			}
			if !found {
				avgListeningSockets = append(avgListeningSockets, &collectorpb.ListeningSocket{
					Command:  socket.Command,
					User:     socket.User,
					Protocol: socket.Protocol,
					Pid:      int64(socket.PID),
					Port:     int64(socket.Port),
				})
			}
		}
		for _, state := range metrics.TCPStates {
			found := false
			for _, avgState := range avgTCPState {
				if avgState.State == state.State {
					avgState.Count += int64(state.Count)
					found = true
					break
				}
			}
			if !found {
				avgTCPState = append(avgTCPState, &collectorpb.TCPStates{
					State: state.State,
					Count: int64(state.Count),
				})
			}
		}

		for _, fs := range metrics.FileSystemUsage {
			found := false
			for _, avgFs := range avgFileSystemUsages {
				if avgFs.FileSystem == fs.FileSystem {
					avgFs.Usedmb += fs.UsedMB
					avgFs.UsedInode += fs.UsedInode
					found = true
					break
				}
			}
			if !found {
				avgFileSystemUsages = append(avgFileSystemUsages, &collectorpb.FileSystemUsage{
					FileSystem: fs.FileSystem,
					Usedmb:     fs.UsedMB / float64(count),
					UsedInode:  fs.UsedInode / float64(count),
				})
			}
		}

		for _, proto := range metrics.NetworkProtocol {
			protocolBytes[proto.Protocol] += proto.Bytes
		}
	}

	networkProtocols := []*collectorpb.NetworkProtocol{}
	for proto, bytes := range protocolBytes {
		networkProtocols = append(networkProtocols, &collectorpb.NetworkProtocol{
			Protocol: proto,
			Bytes:    bytes,
		})
	}

	return &collectorpb.Collector{
		Loadaverage:     avgLoad,
		Cpuusage:        avgCPU,
		Filesystemusage: avgFileSystemUsages,
		Networkprotocol: networkProtocols,
		Trafficinfo:     avgConnections,
		Tcpstates:       avgTCPState,
		Listeningsocket: avgListeningSockets,
		Diskusage:       avgDisk,
	}
}

func StartServer(col *collector.Collector, grpcport string) {
	server := grpc.NewServer()
	collectorpb.RegisterMetricsCollectorServer(server, &MetricsCollectorServer{
		collector: col,
	})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", grpcport))
	if err != nil {
		panic("Failed to listen on port " + grpcport + ": " + err.Error())
	}
	if err := server.Serve(lis); err != nil {
		panic("Failed to serve gRPC server: " + err.Error())
	}
}
