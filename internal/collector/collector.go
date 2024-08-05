package collector

type LoadAverage struct {
	OneMinute      float64
	FiveMinutes    float64
	FifteenMinutes float64
}

type CPUUsage struct {
	UserMode   float64
	SystemMode float64
	Idle       float64
}

type DiskUsage struct {
	Name     string
	TPS      float64
	KBPerSec float64
}

type FileSystemUsage struct {
	FileSystem   string
	UsedMB       float64
	UsedPercent  float64
	UsedInode    float64
	InodePercent float64
}

type NetworkProtocol struct {
	Protocol string
	Bytes    int64
	Percent  float64
}

type ListeningSocket struct {
	Command  string
	PID      int
	User     string
	Protocol string
	Port     int
}

type TCPStates struct {
	State string
	Count int
}

type Collector struct {
	LoadAverage     LoadAverage
	CPUUsage        CPUUsage
	DiskUsage       []DiskUsage
	FileSystemUsage []FileSystemUsage
	NetworkProtocol []NetworkProtocol
	TrafficInfo     []TrafficInfo
	TCPStates       []TCPStates
	ListeningSocket []ListeningSocket
}

func Collect() *Collector {
	var (
		loadAvg          LoadAverage
		cpuUsage         CPUUsage
		diskUsage        []DiskUsage
		fileSystemUsage  []FileSystemUsage
		networkProtocols []NetworkProtocol
		trafficInfo      []TrafficInfo
		tcpStates        []TCPStates
		listeningSocket  []ListeningSocket
	)

	loadAvg, _ = LoadAvg()
	cpuUsage, _ = CpuStat()
	diskUsage, _ = DiskStat()
	fileSystemUsage = FsStat()
	networkProtocols, trafficInfo, tcpStates, listeningSocket = TrafficGetInfo()

	return &Collector{
		LoadAverage:     loadAvg,
		CPUUsage:        cpuUsage,
		DiskUsage:       diskUsage,
		FileSystemUsage: fileSystemUsage,
		NetworkProtocol: networkProtocols,
		TCPStates:       tcpStates,
		TrafficInfo:     trafficInfo,
		ListeningSocket: listeningSocket,
	}
}
