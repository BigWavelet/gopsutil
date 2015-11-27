package cpu

import (
	"encoding/json"
	"runtime"
)

// Documentation
// linux: https://www.kernel.org/doc/Documentation/filesystems/proc.txt  1.8
type CPUTimesStat struct {
	CPU       string `json:"cpu"`
	User      uint64 `json:"user"`
	System    uint64 `json:"system"`
	Idle      uint64 `json:"idle"`
	Nice      uint64 `json:"nice"`
	Iowait    uint64 `json:"iowait"`
	Irq       uint64 `json:"irq"`
	Softirq   uint64 `json:"softirq"`
	Steal     uint64 `json:"steal"`
	Guest     uint64 `json:"guest"`
	GuestNice uint64 `json:"guest_nice"`
	Stolen    uint64 `json:"stolen"`
}

type CPUInfoStat struct {
	CPU        int32    `json:"cpu"`
	VendorID   string   `json:"vendor_id"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Stepping   int32    `json:"stepping"`
	PhysicalID string   `json:"physical_id"`
	CoreID     string   `json:"core_id"`
	Cores      int32    `json:"cores"`
	ModelName  string   `json:"model_name"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int32    `json:"cache_size"`
	Flags      []string `json:"flags"`
}

var lastCPUTimes []CPUTimesStat
var lastPerCPUTimes []CPUTimesStat

func CPUCounts(logical bool) (int, error) {
	return runtime.NumCPU(), nil
}

func (c CPUTimesStat) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func (c CPUInfoStat) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}

func init() {
	lastCPUTimes, _ = CPUTimes(false)
	lastPerCPUTimes, _ = CPUTimes(true)
}
