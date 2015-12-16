package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
	NumCPU  int    `json:"ncpu"`
	Root    bool   `json:"root"`
	Sdk     int    `json:"sdk"`
	Version string `json:"version"`
}

func DumpAndroidInfo() {
	sdk, _ := AndroidSdkVersion()
	ver, _ := getprop("ro.build.version.release")
	di := &DeviceInfo{
		NumCPU:  cpu.CPUCount,
		Root:    IsAndroidRoot(),
		Sdk:     sdk,
		Version: ver,
	}
	data, _ := json.MarshalIndent(di, "", "    ")
	fmt.Println(string(data))
}

// check if root
func IsAndroidRoot() bool {
	for _, searchDir := range []string{"/system/bin/", "/system/xbin/", "/system/sbin/", "/sbin/", "/vendor/bin/"} {
		if fileExists(filepath.Join(searchDir, "su")) {
			return true
		}
	}
	return false
}

// android sdk verion
// http://netease.github.io/airtest/wikipedia/api-version.html
func AndroidSdkVersion() (ver int, err error) {
	val, err := getprop("ro.build.version.sdk")
	fmt.Sscanf(val, "%d", &ver)
	return
}

/*
func AndroidScreenSize() (width int, height int, err error) {
	out, err := exec.Command("dumpsys", "window").Output()
	if err != nil {
		return
	}
	rsRE := regexp.MustCompile(`\s*mRestrictedScreen=\(\d+,\d+\) (?P<w>\d+)x(?P<h>\d+)`)
	matches := rsRE.FindStringSubmatch(string(out))
	if len(matches) == 0 {
		err = errors.New("get shape(width,height) from device error")
		return
	}
	return atoi(matches[1]), atoi(matches[2]), nil
}
*/

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
	//println("FPS")
	stdout := bytes.NewBuffer(nil)
	//sh := exec.Command("su", "-c", "/system/bin/sh")
	sh = exec.Command("su")
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
				//default:
				//println("FULL")
				//}
			}
			lastframe = curframe
			lasttime = time.Now()
			time.Sleep(time.Millisecond * 500)
		}
	}()
	return
}
