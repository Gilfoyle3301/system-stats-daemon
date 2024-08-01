package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func CpuStat() (CPUUsage, error) {
	var objectStat CPUUsage

	stat, err := os.Open("/proc/stat")
	if err != nil {
		return objectStat, fmt.Errorf("failed to open /proc/stat: %w", err)
	}
	scanner := bufio.NewScanner(stat)

	if !scanner.Scan() {
		err := scanner.Err().Error()
		return objectStat, fmt.Errorf("failed to read /proc/stat: %w", err)
	}
	parseField := strings.Fields(scanner.Text())
	userTime, err := strconv.ParseInt(parseField[1], 10, 64)
	if err != nil {
		return objectStat, fmt.Errorf("failed to parse /proc/stat: %w", err)
	}
	niceTime, err := strconv.ParseInt(parseField[2], 10, 64)
	if err != nil {
		return objectStat, fmt.Errorf("failed to parse /proc/stat: %w", err)
	}
	systemTime, err := strconv.ParseInt(parseField[3], 10, 64)
	if err != nil {
		return objectStat, fmt.Errorf("failed to parse /proc/stat: %w", err)
	}
	idleTime, err := strconv.ParseInt(parseField[4], 10, 64)
	if err != nil {
		return objectStat, fmt.Errorf("failed to parse /proc/stat: %w", err)
	}
	return CPUUsage{
		UserMode:   float64(userTime + niceTime),
		SystemMode: float64(systemTime),
		Idle:       float64(idleTime),
	}, nil
}
