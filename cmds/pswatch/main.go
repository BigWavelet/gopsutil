package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/codeskyblue/gopsutil/cpu"
	"github.com/codeskyblue/gopsutil/mem"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Data struct {
	Time int64       `json:"time"`
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

var outC chan *Data

func collectCPU() (*Data, error) {
	v, err := cpu.CPUPercent(0, false)
	if err != nil {
		return nil, err
	}
	cpu := v[0]
	return &Data{
		Time: time.Now().UnixNano() / 1e6,
		Name: "cpu",
		Data: cpu,
	}, nil
}

func collectMem() (*Data, error) {
	v, err := mem.VirtualMemory()
	return &Data{
		Time: time.Now().UnixNano() / 1e6,
		Name: "mem",
		Data: map[string]uint64{
			"total": v.Total,
			"free":  v.Free,
			"used":  v.Used,
		},
	}, err
}

func collectData() {
	wg := sync.WaitGroup{}
	cronCollect := func(collec func() (*Data, error), interval time.Duration) {
		wg.Add(1)
		go func() {
			for {
				start := time.Now()
				data, err := collec()
				if err == nil {
					outC <- data
				}
				spend := time.Since(start)
				if interval > spend {
					time.Sleep(interval - spend)
				}
			}
		}()
	}

	cronCollect(collectCPU, time.Second)
	cronCollect(collectMem, time.Second)
	wg.Wait()
}

func main() {
	outC = make(chan *Data, 5)
	go func() {
		for data := range outC {
			dataByte, _ := json.Marshal(data)
			fmt.Println(string(dataByte))
		}
	}()
	collectData()
}
