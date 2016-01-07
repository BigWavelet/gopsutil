package main

import (
	"fmt"
	"os/exec"
	"regexp"
)

// ref: http://android-test-tw.blogspot.jp/2012/10/dumpsys-information-android-open-source.html

const (
	BATTERY_STATUS_UNKNOWN      = 1
	BATTERY_STATUS_CHARGING     = 2
	BATTERY_STATUS_DISCHARGING  = 3
	BATTERY_STATUS_NOT_CHARGING = 4
	BATTERY_STATUS_FULL         = 5

	BATTERY_HEALTH_UNKNOWN             = 1
	BATTERY_HEALTH_GOOD                = 2
	BATTERY_HEALTH_OVERHEAT            = 3
	BATTERY_HEALTH_DEAD                = 4
	BATTERY_HEALTH_OVER_VOLTAGE        = 5
	BATTERY_HEALTH_UNSPECIFIED_FAILURE = 6
	BATTERY_HEALTH_COLD                = 7
)

type Battery struct {
	ACPowered       bool
	USBPowered      bool
	WirelessPowered bool
	Status          int
	Health          int
	Present         bool
	Level           int
	Scale           int
	Voltage         int
	Temperature     int
	Technology      string
}

func (self *Battery) StatusName() string {
	var ss = []string{"unknown", "charging", "discharging", "notcharging", "full"}
	if self.Status < 1 || self.Status > 5 {
		return "unknown"
	}
	return ss[self.Status-1]
}

func dumpsysCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("dumpsys")
	cmd.Args = append(cmd.Args, args...)
	return cmd.Output()
}

func parseBool(s string) bool {
	return s == "true"
}

func parseInt(s string) int {
	var a int
	fmt.Sscanf(s, "%d", &a)
	return a
}

func (self *Battery) Update() error {
	out, err := dumpsysCommand("battery")
	if err != nil {
		return err
	}
	//log.Println(string(out))
	patten := regexp.MustCompile(`(\w+[\w ]*\w+):\s*([-\w\d]+)(\r|\n)`)
	ms := patten.FindAllStringSubmatch(string(out), -1)
	for _, fields := range ms {
		var key, val = fields[1], fields[2]
		switch key {
		case "AC powered":
			self.ACPowered = parseBool(val)
		case "USB powered":
			self.USBPowered = parseBool(val)
		case "Wireless powered":
			self.WirelessPowered = parseBool(val)
		case "status":
			self.Status = parseInt(val)
		case "present":
			self.Present = parseBool(val)
		case "level":
			self.Level = parseInt(val)
		case "scale":
			self.Scale = parseInt(val)
		case "voltage":
			self.Voltage = parseInt(val)
		case "temperature":
			self.Temperature = parseInt(val)
		case "technology":
			self.Technology = val
		}
	}
	return nil
}
