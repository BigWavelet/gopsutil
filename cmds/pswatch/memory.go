package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/codeskyblue/gopsutil/process"
)

/* need root, unit Byte */
type MemoryStat struct {
	VSS uint64 `json:"vss"`
	RSS uint64 `json:"rss"`
	PSS uint64 `json:"pss"`
	USS uint64 `json:"-"`
}

func isRootUser() bool {
	return os.Getuid() == 0
}

func ProcessMemory(proc *process.Process) (ms *MemoryStat, err error) {
	return StupidGetMemory(proc)

	// Not working in some machines
	if !isRootUser() {
		return StupidGetMemory(proc)
	}

	mapsStat, err := proc.MemoryMaps(true) // need root here
	if err != nil {
		return nil, err
	}
	ms = &MemoryStat{}
	for _, mstat := range *mapsStat {
		ms.VSS += (mstat.Size << 10)
		ms.RSS += (mstat.Rss << 10)
		ms.PSS += (mstat.Pss << 10)
		if mstat.Rss == mstat.Pss {
			ms.USS += (mstat.Rss << 10)
		} else {
			ms.USS += ((mstat.PrivateClean + mstat.PrivateDirty) << 10)
		}
	}
	return
}

var (
	ptn1 = regexp.MustCompile(`PSS:(\s+\d+)+`)
	ptn2 = regexp.MustCompile(`TOTAL\s+(\d+)`)
)

func StupidGetMemory(proc *process.Process) (ms *MemoryStat, err error) {
	ms = &MemoryStat{}
	memStat, err := proc.MemoryInfoEx()
	if err != nil {
		return nil, err
	}
	ms.VSS = memStat.VMS
	ms.RSS = memStat.RSS

	// PSS from dumpsys
	appname, _ := proc.Cmdline()
	cmd := exec.Command("/system/bin/dumpsys", "meminfo", appname)
	if data, er := cmd.CombinedOutput(); er == nil {
		res1 := ptn1.FindStringSubmatch(string(data))
		if len(res1) != 0 {
			fmt.Sscanf(res1[1], "%d", &ms.PSS)
		} else {
			res2 := ptn2.FindStringSubmatch(string(data))
			if len(res2) == 0 {
				return ms, nil
			}
			fmt.Sscanf(res2[1], "%d", &ms.PSS)
		}
		ms.PSS <<= 10
	}
	return ms, nil
}
