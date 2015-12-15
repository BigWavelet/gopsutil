package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/codeskyblue/gopsutil/process"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Data struct {
	Time int64       `json:"-"`
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type CollectFunc func() (*Data, error)

var (
	outC         chan *Data
	collectFuncs = []CollectFunc{}
)

var (
	search = flag.String("p", "",
		"search process, support ex: pid:71, exe:/usr/bin/ls, cmdline:./ps")
)

func main() {
	flag.Parse()

	var proc *process.Process
	if *search != "" {
		procs, err := FindProcess(*search)
		if err != nil {
			log.Fatal(err)
		}
		if len(procs) == 0 {
			log.Fatalf("No process found by %s", strconv.Quote(*search))
		}
		if len(procs) > 1 {
			log.Fatal("Find more then one process matched, This is a bug, maybe")
		}
		proc = procs[0]
		collectFuncs = append(collectFuncs, NewProcCollectCPU(proc))
	}

	outC = make(chan *Data, 5)
	drainData()
	for data := range outC {
		dataByte, _ := json.Marshal(data)
		fmt.Println(string(dataByte))
	}
}

func drainData() {
	for _, collect := range collectFuncs {
		goCronCollect(collect, time.Second, outC)
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
