package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/codeskyblue/gopsutil/android"
	"github.com/codeskyblue/gopsutil/cpu"
	"github.com/codeskyblue/gopsutil/mem"
	"github.com/codeskyblue/gopsutil/process"
	humanize "github.com/dustin/go-humanize"
)

type Data struct {
	Time int64       `json:"time"`
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type CollectFunc func() (*Data, error)
type CollectUnit struct {
	Func     CollectFunc
	Duration time.Duration
}

var (
	firstCpu     = true
	outC         = make(chan *Data, 5)
	collectFuncs = map[string]CollectUnit{} //CollectFunc]time.Duration{} //[]CollectFunc{}
)

func drainData(collects map[string]CollectUnit) {
	for _, cu := range collects {
		goCronCollect(cu.Func, cu.Duration, outC)
	}
}

func goCronCollect(collec CollectFunc, interval time.Duration, outC chan *Data) chan bool {
	done := make(chan bool, 0)
	go func() {
		for {
			start := time.Now()
			data, err := collec()
			if err == nil {
				outC <- data
			} else {
				outC <- &Data{
					Name: "error",
					Data: err.Error(),
				}
			}
			spend := time.Since(start)
			if interval > spend {
				time.Sleep(interval - spend)
			}
		}
		done <- true
	}()
	return done
}

func collectCPU() (*Data, error) {
	if firstCpu {
		time.Sleep(500 * time.Millisecond)
		firstCpu = false
	}

	v, err := cpu.CPUPercent(0, false)
	if err != nil {
		return nil, err
	}
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

func collectBattery() (*Data, error) {
	var bt = Battery{}
	if err := bt.Update(); err != nil {
		return nil, err
	}
	return &Data{
		Name: "battery",
		Data: map[string]interface{}{
			"voltage":      bt.Voltage,
			"temperature":  bt.Temperature,
			"powerPercent": bt.Level * 100 / bt.Scale,
		},
	}, nil
}

func collectFPS() (*Data, error) {
	fps, err := android.FPS()
	if err != nil {
		return nil, err
	}
	return &Data{
		Name: "fps",
		Data: fps,
	}, nil
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

func NewProcCollectMemory(proc *process.Process) CollectFunc {
	return func() (*Data, error) {
		mm, err := ProcessMemory(proc)
		if err != nil {
			return nil, err
		}
		return &Data{
			Name: fmt.Sprintf("proc:%d:memory", proc.Pid),
			Data: mm,
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

func serveHTTP(port int) error {
	http.HandleFunc("/api/v1/perf", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST Method is allowed", http.StatusForbidden)
			return
		}
		name := r.FormValue("name")
		value := r.FormValue("data")
		var v float64
		var data *Data
		if _, err := fmt.Sscanf(value, "%f", &v); err != nil {
			data = &Data{
				Name: "error",
				Data: fmt.Sprintf("http /api/v1/perf read not float data, name: %s, val: %s",
					name, value),
			}
		} else {
			data = &Data{
				Name: name,
				Data: v,
			}
		}
		outC <- data
		w.Header().Set("Content-Type", "application/json")
		jsonData, _ := json.Marshal(data)
		w.Write(jsonData)
	})
	return http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
