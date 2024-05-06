//go:build windows
// +build windows

package lib

import (
	"fmt"
	"os"
	"syscall"
)

func GetStringFromInfo(dir os.DirEntry) string {
	i, _ := dir.Info()
	info := i.Sys().(*syscall.Win32FileAttributeData)
	return fmt.Sprintf("%d%d%d%d%d", info.FileAttributes, info.CreationTime.Nanoseconds(), info.LastWriteTime.Nanoseconds(), info.FileSizeHigh, info.FileSizeLow)
}
