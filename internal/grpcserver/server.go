package grpcserver

import (
	"sync"
	"time"

	collectorpb "github.com/Gilfoyle3301/system-stats-daemon/api/pb"
	"github.com/Gilfoyle3301/system-stats-daemon/internal/collector"
)

type MetricsCollectorServer struct {
	collectorpb.UnimplementedMetricsCollectorServer
}

func (s *MetricsCollectorServer) CollectMetrics(req *collectorpb.MetricsRequest, stream collectorpb.MetricsCollector_CollectMetricsServer) error {
	period := time.Duration(req.GetNSecond()) * time.Second
	averageTime := time.Duration(req.GetMSecond()) * time.Second
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	var (
		dataList []*collector.Collector
		mu       sync.Mutex
	)
	for {
		select {
		case <-ticker.C:
			data := collector.Collect()
			mu.Lock()
			dataList = append(dataList, data)
			if len(dataList) < int(averageTime/period) {
				dataList = dataList[1:]
			}
			mu.Unlock()

			averageData := computeAverages(dataList)
			response := collectorpb.MetricsResponse{
				Collector: averageData,
			}
			if err := stream.Send(&response); err != nil {
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
	avgDisk := &collectorpb.DiskUsage{}
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

		for _, conn := range metrics.TrafficInfo {
			found := false
			for _, accConn := range avgConnections {
				if accConn.Sourceip == conn.SourceIP && int(accConn.SourcePort) == conn.SourcePort &&
					accConn.Destip == conn.DestIP && int(accConn.DestPort) == conn.DestPort &&
					accConn.Protocol == conn.Protocol && accConn.State == conn.State {
					accConn.Bytes += int64(conn.Bytes)
					accConn.Bps += conn.BPS
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

		for _, state := range metrics.TCPStates {
			found := false
			for _, accState := range avgTCPState {
				if accState.State == state.State {
					accState.Count += int64(state.Count)
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

		avgListeningSockets = append(avgListeningSockets, &collectorpb.ListeningSocket{})

		for _, fs := range metrics.FileSystemUsage {
			found := false
			for _, accFS := range avgFileSystemUsages {
				if accFS.FileSystem == fs.FileSystem {
					accFS.Usedmb += fs.UsedMB
					accFS.UsedInode += fs.UsedInode
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
			if _, exists := protocolBytes[proto.Protocol]; exists {
				protocolBytes[proto.Protocol] += proto.Bytes
			} else {
				protocolBytes[proto.Protocol] = proto.Bytes
			}
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
