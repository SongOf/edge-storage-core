package system

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	fmt.Println(cpu.Counts(false))
	fmt.Println(cpu.Counts(true))
	fmt.Println(cpu.Info())
	fmt.Println(cpu.Times(false))
	fmt.Println(cpu.Percent(time.Second, false))

	fmt.Println(load.Avg())

	fmt.Println(disk.Partitions(true))
	fmt.Println(disk.Usage("/etc"))

	fmt.Println(mem.VirtualMemory())
	fmt.Println(mem.SwapMemory())

	fmt.Println(net.IOCounters(true))
}
