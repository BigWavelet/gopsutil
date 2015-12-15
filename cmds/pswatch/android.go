package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"time"
)

// android sdk verion
// http://netease.github.io/airtest/wikipedia/api-version.html
func AndroidSdkVersion() (ver int, err error) {
	c := exec.Command("getprop", "ro.build.version.sdk")
	data, err := c.Output()
	if err != nil {
		return
	}
	fmt.Sscanf(string(data), "%d", &ver)
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
