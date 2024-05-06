//go:build !windows && (arm64 || amd64)
// +build !windows
// +build arm64 amd64

package lib

import (
	"crypto/md5"
	"fmt"
	"os"
	"syscall"
)

func GetStringFromInfo(dir os.DirEntry) string {
	hash := md5.New()
	info, _ := dir.Info()
	stat := info.Sys().(*syscall.Stat_t)
	return string(hash.Sum([]byte(fmt.Sprint(stat.Size, stat.Mtimespec, stat.Mode, stat.Dev, stat.Ino))))
}
