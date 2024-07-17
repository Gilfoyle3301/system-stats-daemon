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

type TrafficInfo struct {
	SourceIP   string
	SourcePort int
	DestIP     string
	DestPort   int
	Protocol   string
	BPS        float64
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
	DiskUsage       DiskUsage
	FileSystemUsage []FileSystemUsage
	NetworkProtocol []NetworkProtocol
	TrafficInfo     []TrafficInfo
	TCPStates       []TCPStates
	ListeningSocket []ListeningSocket
}

func New() *Collector {
	return &Collector{}
}

// func collectLoadAverage() LoadAverage {

// }

// func collectCPUUsage() CPUUsage {

// }

// func collectDiskUsage() DiskUsage {

// }

// func collectFileSystemUsage() []FileSystemUsage {

// }

// func collectNetworkProtocol() []NetworkProtocol {

// }

// func collectTCPStates() []TCPStates {

// }

// func collectListeningSocket() []ListeningSocket {

// }
