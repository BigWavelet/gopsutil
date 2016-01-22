package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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
		Root:    IsAndroidRoot(),
		Sdk:     sdk,
		Version: ver,
	}
	data, _ := json.MarshalIndent(di, "", "    ")
	fmt.Println(string(data))
}

// check if android rooted
func IsAndroidRoot() bool {
	paths := strings.Split(os.Getenv("PATH"), ":")
	paths = append(paths, "/system/bin/", "/system/xbin/", "/system/sbin/", "/sbin/", "/vendor/bin/")
	for _, searchDir := range UniqSlice(paths) {
		suPath := filepath.Join(searchDir, "su")
		suStat, err := os.Lstat(suPath)
		if err == nil && suStat.Mode().IsRegular() {
			// check if setuid is set
			if suStat.Mode()&os.ModeSetuid == os.ModeSetuid {
				return true
			}
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

func readUint64FromFile(filename string) (uint64, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}

// ref
// traffix for android: http://keepcleargas.bitbucket.org/2013/10/12/android-App-Traffic.html
// get uid from /proc/<pid>/status http://www.linuxquestions.org/questions/linux-enterprise-47/uid-and-gid-fileds-from-proc-pid-status-595383/
func ReadTrafix(uid int32) (rcv, snd uint64, err error) {
	tcpRecv := fmt.Sprintf("/proc/uid_stat/%d/tcp_rcv", uid)
	tcpSend := fmt.Sprintf("/proc/uid_stat/%d/tcp_snd", uid)
	rcv, err = readUint64FromFile(tcpRecv)
	if err != nil {
		return
	}
	snd, err = readUint64FromFile(tcpSend)
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
