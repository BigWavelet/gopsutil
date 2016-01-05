package main

import (
	"fmt"

	"github.com/codeskyblue/gopsutil/cpu"
	"github.com/codeskyblue/gopsutil/mem"
	"github.com/codeskyblue/gopsutil/process"
	humanize "github.com/dustin/go-humanize"
)

func init() {
	collectFuncs = append(collectFuncs,
		collectCPU, collectMem)
}

func collectCPU() (*Data, error) {
	//v, err := cpu.CPUPercent(0, true)
	v, err := cpu.CPUPercent(0, false)
	if err != nil {
		return nil, err
	}
	/*
		var total float64
		for _, val := range v {
			total += val
		}
		avg := total / float64(len(v))
	*/
	avg := v[0]
	return &Data{
		Name: "cpu",
		Data: map[string]interface{}{
			//"count":   len(v),
			//"each":    v,
			"average": avg,
			"summary": fmt.Sprintf("Average CPU Percent: %.2f%%", avg),
		},
	}, nil
}

func collectMem() (*Data, error) {
	v, err := mem.VirtualMemory()
	return &Data{
		Name: "mem",
		Data: map[string]interface{}{
			"total": v.Total,
			"free":  v.Free,
			"used":  v.Used,
			"summary": fmt.Sprintf("Total: %v, Free %v, Used %v",
				humanize.IBytes(v.Total), humanize.IBytes(v.Free), humanize.IBytes(v.Used)),
		},
	}, err
}

// search name accept some process
// name:<short name>
// pid:<pid>
// cmdline:<name>  Find android app can use this way
func FindProcess(search string) (procs []*process.Process, err error) {
	pids, err := process.Pids()
	if err != nil {
		return
	}
	procs = make([]*process.Process, 0)
	for _, pid := range pids {
		proc, er := process.NewProcess(pid)
		if er != nil {
			continue
		}
		if isMatch(proc, search) {
			procs = append(procs, proc)
		}
	}
	//return nil, errors.New("Process not found by search text: " + search)
	return
}

func isMatch(proc *process.Process, search string) bool {
	cmdline, _ := proc.Cmdline()
	if fmt.Sprintf("pid:%d", proc.Pid) == search {
		return true
	}
	if "cmdline:"+cmdline == search {
		return true
	}
	exe, _ := proc.Exe()
	if "exe:"+exe == search {
		return true
	}
	name, _ := proc.Name()
	if "name:"+name == search {
		return true
	}
	return false
}

func NewProcCollectCPU(proc *process.Process) CollectFunc {
	return func() (*Data, error) {
		percent, err := proc.CPUPercent(0)
		if err != nil {
			return nil, err
		}
		return &Data{
			Name: fmt.Sprintf("proc:%d:cpu", proc.Pid),
			Data: map[string]interface{}{
				"total":   percent,
				"average": percent / float64(cpu.CPUCount),
			},
		}, nil
	}
}

func NewProcCollectTraffic(proc *process.Process) CollectFunc {
	return func() (*Data, error) {
		uids, err := proc.Uids()
		if err != nil {
			return nil, err
		}
		uid := uids[0] // there are four, first is real_uid
		rcv, snd, err := ReadTrafix(uid)
		if err != nil {
			return nil, err
		}
		return &Data{
			Name: fmt.Sprintf("proc:%d:traffic", proc.Pid),
			Data: map[string]interface{}{
				"recv": rcv,
				"send": snd,
			},
		}, nil
	}
}

func drainAndroidFPS(outC chan *Data) error {
	sh, pipe, err := drainFPS()
	if err != nil {
		return err
	}
	defer func() {
		if sh.Process != nil {
			sh.Process.Kill()
		}
	}()
	for val := range pipe {
		//log.Println("FPS:", val)
		outC <- &Data{
			Name: "fps",
			Data: val,
		}
	}
	return nil
}
