package collector

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func LoadAvg() (LoadAverage, error) {
	var objectLA LoadAverage

	loadavg, err := os.Open("/proc/loadavg")
	if err != nil {
		return objectLA, fmt.Errorf("failed to open /proc/loadavg: %w", err)
	}
	defer loadavg.Close()

	scanner := bufio.NewScanner(loadavg)
	if !scanner.Scan() {
		err := scanner.Err()
		return objectLA, fmt.Errorf("failed to read /proc/loadavg: %w", err)
	}
	line := strings.Fields(scanner.Text())
	OneMinute, err := strconv.ParseFloat(line[0], 64)
	if err != nil {
		return objectLA, fmt.Errorf("failed to parse value /proc/loadavg: %w", err)
	}
	FiveMinutes, err := strconv.ParseFloat(line[1], 64)
	if err != nil {
		return objectLA, fmt.Errorf("failed to parse value /proc/loadavg: %w", err)
	}
	FifteenMinutes, err := strconv.ParseFloat(line[2], 64)
	if err != nil {
		return objectLA, fmt.Errorf("failed to parse value /proc/loadavg: %w", err)
	}

	return LoadAverage{
		OneMinute:      float64(OneMinute),
		FiveMinutes:    float64(FiveMinutes),
		FifteenMinutes: float64(FifteenMinutes),
	}, nil
}
