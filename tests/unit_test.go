package collector_test

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"

	collector "github.com/Gilfoyle3301/system-stats-daemon/internal/collector"
	"github.com/stretchr/testify/require"
)

func TestUnitPackage(t *testing.T) {
	t.Run("LoadAverage", func(t *testing.T) {
		testData, err := collector.LoadAvg()
		require.NoError(t, err)
		loadavg, err := os.Open("/proc/loadavg")
		require.NoError(t, err)
		defer loadavg.Close()
		scanner := bufio.NewScanner(loadavg)
		scanner.Scan()
		line := strings.Fields(scanner.Text())
		on, err := strconv.ParseFloat(line[0], 64)
		require.NoError(t, err)
		five, err := strconv.ParseFloat(line[1], 64)
		require.NoError(t, err)
		feev, err := strconv.ParseFloat(line[2], 64)
		require.NoError(t, err)
		require.Equal(t, on, testData.OneMinute)
		require.Equal(t, five, testData.FiveMinutes)
		require.Equal(t, feev, testData.FifteenMinutes)
	})
	t.Run("CPU", func(t *testing.T) {
		testData, err := collector.CpuStat()
		require.NoError(t, err)

		cpu, err := os.Open("/proc/stat")
		require.NoError(t, err)
		defer cpu.Close()

		scanner := bufio.NewScanner(cpu)
		scanner.Scan()
		line := strings.Fields(scanner.Text())
		userTime, err := strconv.ParseInt(line[1], 10, 64)
		require.NoError(t, err)
		niceTime, err := strconv.ParseInt(line[2], 10, 64)
		require.NoError(t, err)
		time := niceTime + userTime
		systemTime, err := strconv.ParseInt(line[3], 10, 64)
		require.NoError(t, err)
		idleTime, err := strconv.ParseInt(line[4], 10, 64)
		require.NoError(t, err)
		require.Equal(t, float64(time), testData.UserMode)
		require.Equal(t, float64(systemTime), testData.SystemMode)
		require.Equal(t, float64(idleTime), testData.Idle)
	})
	t.Run("Disk", func(t *testing.T) {
		testData, err := collector.DiskStat()
		require.NoError(t, err)
		require.NotEmpty(t, testData)
		require.Len(t, testData, len(testData))
	})
	t.Run("trafic", func(t *testing.T) {
		NetworkProtocol, TrafficInfo, TCPStates, ListeningSocket := collector.TrafficGetInfo()
		require.NotEmpty(t, NetworkProtocol)
		require.NotEmpty(t, TrafficInfo)
		require.NotEmpty(t, TCPStates)
		require.NotEmpty(t, ListeningSocket)

	})
	t.Run("filesystem slice", func(t *testing.T) {
		testData := collector.FsStat()
		require.NotEmpty(t, testData)
		require.Len(t, testData, len(testData))
	})
}
