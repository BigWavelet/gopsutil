package android

import (
	"os"
	"path/filepath"
	"strings"
)

// check if android rooted
func IsRoot() bool {
	paths := strings.Split(os.Getenv("PATH"), ":")
	paths = append(paths, "/system/bin/", "/system/xbin/", "/system/sbin/", "/sbin/", "/vendor/bin/")
	paths = uniqSlice(paths)
	for _, searchDir := range paths {
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
