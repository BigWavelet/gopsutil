package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/codeskyblue/gopsutil/process"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Data struct {
	Time int64       `json:"time"`
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type CollectFunc func() (*Data, error)

var (
	outC         = make(chan *Data, 5)
	collectFuncs = []CollectFunc{}
)

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

var (
	VERSION    = "0.0.x"
	BUILD_DATE = "unknown"
)

var (
	search = flag.String("p", "",
		"Search process, support ex: pid:71, exe:/usr/bin/ls, cmdline:./ps")
	showInfo = flag.Bool("i", false, "Show mathine infomation")
	showFPS  = flag.Bool("fps", false, "Show fps of android")
	version  = flag.Bool("v", false, "Show version")
)

func showVersion() error {
	fmt.Printf("version: %v\n", VERSION)
	fmt.Printf("build: %v\n", BUILD_DATE)
	fmt.Printf("golang: %v\n", runtime.Version())
	fd, err := os.Open(os.Args[0])
	if err != nil {
		return err
	}
	md5h := md5.New()
	io.Copy(md5h, fd)
	fmt.Printf("md5sum: %x\n", md5h.Sum([]byte(""))) //md5
	return nil
}

func main() {
	flag.Parse()

	if *version {
		showVersion()
		return
	}

	if *showInfo {
		DumpAndroidInfo()
		return
	}

	if *showFPS {
		go drainAndroidFPS(outC)
	}

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
		collectFuncs = append(collectFuncs,
			NewProcCollectCPU(proc),
			NewProcCollectTraffic(proc))
	}

	drainData()
	for data := range outC {
		data.Time = time.Now().UnixNano() / 1e6 // milliseconds
		dataByte, _ := json.Marshal(data)
		fmt.Println(string(dataByte))
	}
}
