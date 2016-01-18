package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/codeskyblue/gopsutil/process"
)

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
	duration = flag.Duration("d", time.Second, "Collect interval")
	listen   = flag.Bool("l", false, "Listen http request data")
	port     = flag.Int("port", 16118, "Listen port") // because this day is 2016-01-18
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

	if *listen {
		go serveHTTP(*port) // open a server for app to send data
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
		collectFuncs["proc_cpu"] = CollectUnit{NewProcCollectCPU(proc), *duration}
		collectFuncs["proc_net"] = CollectUnit{NewProcCollectTraffic(proc), *duration}
		collectFuncs["proc_mem"] = CollectUnit{NewProcCollectMemory(proc), *duration}
	}
	collectFuncs["sys_cpu"] = CollectUnit{collectCPU, *duration}
	collectFuncs["sys_mem"] = CollectUnit{collectMem, *duration}
	collectFuncs["battery"] = CollectUnit{collectBattery, *duration}

	drainData(collectFuncs)
	for data := range outC {
		data.Time = time.Now().UnixNano() / 1e6 // milliseconds
		dataByte, _ := json.Marshal(data)
		fmt.Println(string(dataByte))
	}
}
