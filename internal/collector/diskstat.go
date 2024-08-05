package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type DiskParams struct {
	rsect uint64
	wsect uint64
	rio   uint64
	wio   uint64
}

func diskCheck() (map[string]DiskParams, error) {
	objectDiskStat := make(map[string]DiskParams)
	diskStat, err := os.Open("/proc/diskstats")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/diskstats: %w", err)
	}
	defer diskStat.Close()
	scanner := bufio.NewScanner(diskStat)
	for scanner.Scan() {
		diskInfo := strings.Fields(scanner.Text())
		rio, err := strconv.ParseInt(diskInfo[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse /proc/diskstats: %w", err)
		}
		rsect, err := strconv.ParseInt(diskInfo[5], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse /proc/diskstats: %w", err)
		}
		wio, err := strconv.ParseInt(diskInfo[7], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse /proc/diskstats: %w", err)
		}
		wsect, err := strconv.ParseInt(diskInfo[9], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse /proc/diskstats: %w", err)
		}
		objectDiskStat[diskInfo[2]] = DiskParams{
			rsect: uint64(rsect),
			wsect: uint64(wsect),
			rio:   uint64(rio),
			wio:   uint64(wio),
		}
	}
	return objectDiskStat, nil
}

func DiskStat() ([]DiskUsage, error) {
	stat := make([]DiskUsage, 50)
	initValue, err := diskCheck()
	if err != nil {
		return nil, err
	}
	time.Sleep(1 * time.Second)
	deltaValue, err := diskCheck()
	if err != nil {
		return nil, err
	}
	for name, data := range initValue {
		tps := (deltaValue[name].wio - data.wio) + (deltaValue[name].rio - data.rio)
		kbPerSec := (deltaValue[name].wsect - data.wsect) + (deltaValue[name].rsect-data.rsect)/1024
		stat = append(stat, DiskUsage{
			Name:     name,
			TPS:      float64(tps),
			KBPerSec: float64(kbPerSec),
		})
	}
	return stat, nil
}
