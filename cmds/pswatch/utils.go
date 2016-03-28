package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func UniqSlice(slice []string) []string {
	found := make(map[string]bool)
	for _, item := range slice {
		found[item] = true
	}
	out := make([]string, 0)
	for key, _ := range found {
		out = append(out, key)
	}
	return out
}

func readUint64FromFile(filename string) (uint64, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
}

func shellEscape(cmds []string) string {
	qcmds := make([]string, 0, len(cmds))
	for _, command := range cmds {
		qcmds = append(qcmds, strconv.Quote(command))
	}
	return strings.Join(qcmds, " ")
}

func rootRun(cmds ...string) error {
	sh := exec.Command("su")
	stdin, _ := sh.StdinPipe()
	sh.Stdout = os.Stdout
	sh.Stderr = os.Stderr
	if err := sh.Start(); err != nil {
		return err
	}

	shcmd := shellEscape(cmds) + "; exit $?\n"
	stdin.Write([]byte(shcmd))
	return sh.Wait()
}
