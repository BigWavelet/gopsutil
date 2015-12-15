// +build linux

package cpu

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/codeskyblue/gopsutil/internal/common"
)

var cpu_tick = float64(100)
var CPUCount = runtime.NumCPU()

func init() {
	out, err := exec.Command("/usr/bin/getconf", "CLK_TCK").Output()
	// ignore errors
	if err == nil {
		i, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		if err == nil {
			cpu_tick = float64(i)
		}
	}
	cpuinfo, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		panic(err)
	}
	patten := regexp.MustCompile(`processor\s*:\s*\d`)
	cpuall := patten.FindAll(cpuinfo, -1)
	CPUCount = len(cpuall)
	// CPUCount, _ = CPUCounts(true)
}

func CPUTimes(percpu bool) ([]CPUTimesStat, error) {
	filename := common.HostProc("stat")
	var lines = []string{}
	if percpu {
		var startIdx uint = 1
		for {
			linen, _ := common.ReadLinesOffsetN(filename, startIdx, 1)
			line := linen[0]
			if !strings.HasPrefix(line, "cpu") {
				break
			}
			lines = append(lines, line)
			startIdx += 1
		}
	} else {
		lines, _ = common.ReadLinesOffsetN(filename, 0, 1)
	}

	ret := make([]CPUTimesStat, 0, len(lines))

	for _, line := range lines {
		ct, err := parseStatLine(line)
		if err != nil {
			continue
		}
		ret = append(ret, *ct)

	}
	return ret, nil
}

func sysCpuPath(cpu int32, relPath string) string {
	return common.HostSys(fmt.Sprintf("devices/system/cpu/cpu%d", cpu), relPath)
}

func finishCPUInfo(c *CPUInfoStat) error {
	if c.Mhz == 0 {
		lines, err := common.ReadLines(sysCpuPath(c.CPU, "cpufreq/cpuinfo_max_freq"))
		if err == nil {
			value, err := strconv.ParseFloat(lines[0], 64)
			if err != nil {
				return err
			}
			c.Mhz = value
		}
	}
	if len(c.CoreID) == 0 {
		lines, err := common.ReadLines(sysCpuPath(c.CPU, "topology/core_id"))
		if err == nil {
			c.CoreID = lines[0]
		}
	}
	return nil
}

// CPUInfo on linux will return 1 item per physical thread.
//
// CPUs have three levels of counting: sockets, cores, threads.
// Cores with HyperThreading count as having 2 threads per core.
// Sockets often come with many physical CPU cores.
// For example a single socket board with two cores each with HT will
// return 4 CPUInfoStat structs on Linux and the "Cores" field set to 1.
func CPUInfo() ([]CPUInfoStat, error) {
	filename := common.HostProc("cpuinfo")
	lines, _ := common.ReadLines(filename)

	var ret []CPUInfoStat

	c := CPUInfoStat{CPU: -1, Cores: 1}
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {
		case "processor":
			if c.CPU >= 0 {
				err := finishCPUInfo(&c)
				if err != nil {
					return ret, err
				}
				ret = append(ret, c)
			}
			c = CPUInfoStat{Cores: 1}
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return ret, err
			}
			c.CPU = int32(t)
		case "vendor_id":
			c.VendorID = value
		case "cpu family":
			c.Family = value
		case "model":
			c.Model = value
		case "model name":
			c.ModelName = value
		case "stepping":
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return ret, err
			}
			c.Stepping = int32(t)
		case "cpu MHz":
			t, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return ret, err
			}
			c.Mhz = t
		case "cache size":
			t, err := strconv.ParseInt(strings.Replace(value, " KB", "", 1), 10, 64)
			if err != nil {
				return ret, err
			}
			c.CacheSize = int32(t)
		case "physical id":
			c.PhysicalID = value
		case "core id":
			c.CoreID = value
		case "flags", "Features":
			c.Flags = strings.FieldsFunc(value, func(r rune) bool {
				return r == ',' || r == ' '
			})
		}
	}
	if c.CPU >= 0 {
		err := finishCPUInfo(&c)
		if err != nil {
			return ret, err
		}
		ret = append(ret, c)
	}
	return ret, nil
}

func parseStatLine(line string) (*CPUTimesStat, error) {
	fields := strings.Fields(line)

	if strings.HasPrefix(fields[0], "cpu") == false {
		//		return CPUTimesStat{}, e
		return nil, errors.New("not contain cpu")
	}

	cpu := fields[0]
	if cpu == "cpu" {
		cpu = "cpu-total"
	}
	puint := func(s string) (uint64, error) {
		return strconv.ParseUint(s, 10, 64)
	}
	user, err := puint(fields[1])
	if err != nil {
		return nil, err
	}
	nice, err := puint(fields[2])
	if err != nil {
		return nil, err
	}
	system, err := puint(fields[3])
	if err != nil {
		return nil, err
	}
	idle, err := puint(fields[4])
	if err != nil {
		return nil, err
	}
	iowait, err := puint(fields[5])
	if err != nil {
		return nil, err
	}
	irq, err := puint(fields[6])
	if err != nil {
		return nil, err
	}
	softirq, err := puint(fields[7])
	if err != nil {
		return nil, err
	}

	ct := &CPUTimesStat{
		CPU:     cpu,
		User:    user,    //float64(user) / cpu_tick,
		Nice:    nice,    //float64(nice) / cpu_tick,
		System:  system,  //float64(system) / cpu_tick,
		Idle:    idle,    //float64(idle) / cpu_tick,
		Iowait:  iowait,  //float64(iowait) / cpu_tick,
		Irq:     irq,     //float64(irq) / cpu_tick,
		Softirq: softirq, //float64(softirq) / cpu_tick,
	}
	if len(fields) > 8 { // Linux >= 2.6.11
		steal, err := puint(fields[8])
		if err != nil {
			return nil, err
		}
		ct.Steal = steal //) / cpu_tick
	}
	if len(fields) > 9 { // Linux >= 2.6.24
		guest, err := puint(fields[9])
		if err != nil {
			return nil, err
		}
		ct.Guest = guest //float64(guest) / cpu_tick
	}
	if len(fields) > 10 { // Linux >= 3.2.0
		guestNice, err := puint(fields[10])
		if err != nil {
			return nil, err
		}
		ct.GuestNice = guestNice //float64(guestNice) / cpu_tick
	}

	return ct, nil
}
