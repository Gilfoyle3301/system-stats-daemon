package collector

import (
	"bufio"
	"log/slog"
	"os"
	"strings"
	"syscall"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func FsStat() []FileSystemUsage {
	var objectFS []FileSystemUsage
	fileFs, err := os.Open("/proc/mounts")
	if err != nil {
		slog.Error(err.Error())
	}
	defer fileFs.Close()
	scanner := bufio.NewScanner(fileFs)

	for scanner.Scan() {
		line := scanner.Text()

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		fs := fields[0]
		mountPoint := fields[1]

		statFs := syscall.Statfs_t{}

		if err := syscall.Statfs(mountPoint, &statFs); err != nil {
			slog.Error(err.Error())
		}
		size := statFs.Blocks * uint64(statFs.Bsize)
		free := statFs.Bfree * uint64(statFs.Bsize)
		used := size - free
		inode := statFs.Files
		inodeFree := statFs.Ffree
		inodeUsed := inode - inodeFree

		var persentUsed float64
		var persentInodUsed float64

		if size > 0 {
			persentUsed = (float64(used) / float64(size)) * 100
			persentInodUsed = (float64(inodeUsed) / float64(inode)) * 100
		}
		objectFS = append(objectFS, FileSystemUsage{
			FileSystem:   fs,
			UsedMB:       float64(used / MB),
			UsedPercent:  float64(persentUsed),
			UsedInode:    float64(inodeUsed),
			InodePercent: persentInodUsed,
		})
	}
	return objectFS
}
