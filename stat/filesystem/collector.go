package filesystem

//collect metrics of local file system usage
import (
	"songof.com/edge-storage-core/pkg/eslog"
	"time"

	"github.com/shirou/gopsutil/disk"
)

type DiskStat struct {
	Path        string
	Fstype      string
	Total       uint64
	Free        uint64
	Used        uint64
	UsedPercent float64
}

// FileSystemStatsHandler represents a handler to handle stats after successfully gathering statistics
type FileSystemStatsHandler func(DiskStat)

// Collector implements the periodic grabbing of informational data of go runtime to a SystemStatsHandler.
type Collector struct {
	// CollectInterval represents the interval in-between each set of stats output.
	// Defaults to 10 seconds.
	CollectInterval time.Duration

	//Path tar file path
	Path string

	// Done, when closed, is used to signal Collector that is should stop collecting
	// statistics and the Run function should return.
	Done <-chan struct{}

	statsHandler FileSystemStatsHandler
}

// New creates a new Collector that will periodically output statistics to statsHandler. It
// will also set the values of the exported stats to the described defaults. The values
// of the exported defaults can be changed at any point before Run is called.
func New(statsHandler FileSystemStatsHandler, path string) *Collector {
	if statsHandler == nil {
		statsHandler = func(DiskStat) {}
	}

	return &Collector{
		CollectInterval: 10 * time.Second,
		Path:            path,
		statsHandler:    statsHandler,
	}
}

// Run gathers statistics then outputs them to the configured SystemStatsHandler every
// CollectInterval. Unlike Once, this function will return until Done has been closed
// (or never if Done is nil), therefore it should be called in its own goroutine.
func (c *Collector) Run() {
	c.statsHandler(c.collectStats())

	tick := time.NewTicker(c.CollectInterval)
	defer tick.Stop()
	for {
		select {
		case <-c.Done:
			return
		case <-tick.C:
			c.statsHandler(c.collectStats())
		}
	}
}

// Once returns a map containing all statistics. It is safe for use from multiple go routines。
func (c *Collector) Once() DiskStat {
	return c.collectStats()
}

// collectStats collects all configured stats once.
func (c *Collector) collectStats() DiskStat {
	s, err := disk.Usage(c.Path)
	if err != nil {
		eslog.L().Error("Failed to get disk usage", eslog.Field("path", c.Path), eslog.Field("Error", err))
	}

	return DiskStat{
		Path:        c.Path,
		Fstype:      s.Fstype,
		Total:       s.Total,
		Free:        s.Free,
		Used:        s.Used,
		UsedPercent: s.UsedPercent,
	}
}

// Values returns metrics which you can write into TSDB.
func (ds *DiskStat) Values() map[string]interface{} {
	values := map[string]interface{}{
		"file.path":        ds.Path,
		"file.fstype":      ds.Fstype,
		"file.total":       ds.Total,
		"file.free":        ds.Free,
		"file.used":        ds.Used,
		"file.usedpercent": ds.UsedPercent,
	}

	return values
}
