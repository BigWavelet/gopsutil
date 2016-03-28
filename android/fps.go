package android

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"time"
)

var (
	surfaceFlingerPatten        = regexp.MustCompile(`\(([a-fA-F0-9]+)\ `)
	lastFrameCount       uint64 = 0
	lastFrameTime               = time.Now()
)

func FPS() (v float64, err error) {
	fcnt, err := frameCount()
	if err != nil {
		return
	}
	if lastFrameCount < 1 {
		lastFrameTime = time.Now()
		time.Sleep(500 * time.Millisecond)
		lastFrameCount = fcnt
		fcnt, _ = frameCount()
	}

	// frame / (time(ms)/1000.0)
	duration := time.Since(lastFrameTime)
	v = float64(fcnt-lastFrameCount) / (float64(duration.Nanoseconds()/1e6) / 1000.0)
	lastFrameTime = time.Now()
	lastFrameCount = fcnt
	return
}

func frameCount() (count uint64, err error) {
	cmd := exec.Command("service", "call", "SurfaceFlinger", "1013")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	result := surfaceFlingerPatten.FindStringSubmatch(string(out))
	if len(result) < 2 {
		err = errors.New(string(out))
		return
	}
	_, err = fmt.Sscanf(result[1], "%x", &count)
	return

	/*
		//func drainFPS() (sh *exec.Cmd, pipe chan float64, err error) {
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
	*/
}
