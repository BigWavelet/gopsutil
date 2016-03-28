package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/codeskyblue/gopsutil/android"
	"github.com/codeskyblue/gopsutil/cpu"
)

func atoi(a string) int {
	var i int
	_, err := fmt.Sscanf(a, "%d", &i)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getprop(key string) (string, error) {
	c := exec.Command("getprop", key)
	data, err := c.Output()
	return strings.TrimSpace(string(data)), err
}

type DeviceInfo struct {
	Serial  string `json:"serial"`
	NumCPU  int    `json:"ncpu"`
	Root    bool   `json:"root"`
	Sdk     int    `json:"sdk"`
	Version string `json:"version"`
}

func DumpAndroidInfo() {
	sdk, _ := AndroidSdkVersion()
	ver, _ := getprop("ro.build.version.release")
	serial, _ := getprop("ro.serialno")
	di := &DeviceInfo{
		Serial:  serial,
		NumCPU:  cpu.CPUCount,
		Root:    android.IsRoot(),
		Sdk:     sdk,
		Version: ver,
	}
	data, _ := json.MarshalIndent(di, "", "    ")
	fmt.Println(string(data))
}

// android sdk verion
// http://netease.github.io/airtest/wikipedia/api-version.html
func AndroidSdkVersion() (ver int, err error) {
	val, err := getprop("ro.build.version.sdk")
	fmt.Sscanf(val, "%d", &ver)
	return
}

// ref
// get uid from /proc/<pid>/status http://www.linuxquestions.org/questions/linux-enterprise-47/uid-and-gid-fileds-from-proc-pid-status-595383/
func ReadTrafix(uid int32) (rcv, snd uint64, err error) {
	nss, err := android.NetworkStats()
	if err != nil {
		return
	}
	rcv = 0
	snd = 0
	for _, ns := range nss {
		if ns.Uid != int(uid) {
			continue
		}
		//rcv += ns.RecvTcpBytes - 64*ns.RecvTcpPackets
		//snd += ns.SendTcpBytes - 64*ns.SendTcpPackets
		rcv += ns.RecvBytes
		snd += ns.SendBytes
	}
	return
}

var patten = regexp.MustCompile(`\(([a-fA-F0-9]+)\ `)

func drainFPS() (sh *exec.Cmd, pipe chan float64, err error) {
	pipe = make(chan float64)
	ver, err := AndroidSdkVersion()
	if err != nil {
		return
	}
	if ver < 14 {
		err = fmt.Errorf("andorid sdk version should >=14, but got %d\n", ver)
		return
	}
	var lastframe = 0
	var lasttime = time.Now()
	stdout := bytes.NewBuffer(nil)
	sh = exec.Command("su") // exec.Command("su", "-c", "/system/bin/sh")
	stdin, _ := sh.StdinPipe()
	sh.Stdout = stdout
	sh.Stderr = stdout
	if err = sh.Start(); err != nil {
		log.Println("run command failed: su")
		return
	}
	go func() {
		for {
			stdin.Write([]byte("service call SurfaceFlinger 1013\n"))
			var err error
			var line string
			for {
				line, err = stdout.ReadString('\n')
				if err != nil && err != io.EOF {
					log.Fatal(err)
				}
				if err == io.EOF {
					time.Sleep(time.Millisecond * 10)
					continue
				}
				break
			}
			result := patten.FindStringSubmatch(line)
			var curframe int
			fmt.Sscanf(result[1], "%x", &curframe)
			if lastframe != 0 {
				fpsrate := float64(curframe-lastframe) / time.Now().Sub(lasttime).Seconds()
				//fmt.Printf("%.1f\n", fpsrate)
				//select {
				//case pipe <- fpsrate:
				pipe <- fpsrate
			}
			lastframe = curframe
			lasttime = time.Now()
			time.Sleep(time.Millisecond * 500)
		}
	}()
	return
}
