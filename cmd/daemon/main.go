package main

import (
	"fmt"

	"github.com/Gilfoyle3301/system-stats-daemon/internal/collector"
)

func main() {
	// newCollect := collector.New()
	// fmt.Println(newCollect)
	// FS := collector.DiskFree()
	// for _, ob := range FS {
	// 	fmt.Println(ob.FileSystem, ob.UsedMB, ob.UsedPercent, ob.UsedInode, ob.InodePercent)
	// }
	avg, _ := collector.DiskStat()
	fmt.Println(avg)

}
