package main

import (
	"io/ioutil"
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
