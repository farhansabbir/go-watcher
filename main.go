package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type ignores []string

var (
	watch_location    string            = "./"
	watch_delay_milli int               = 100
	command           string            = ""
	entries           map[string]string = make(map[string]string)
	skips             ignores           = ignores([]string{".git", "node_modules"})
)

func (i *ignores) Set(value string) error {
	for _, str := range strings.Split(value, ",") {
		*i = append(*i, str)
	}
	return nil
}

func (i *ignores) String() string {
	b, _ := json.Marshal(*i)
	return string(b)
}

func init() {
	flag.StringVar(&watch_location, "watch", "./", "location to watch for changes.")
	flag.IntVar(&watch_delay_milli, "delay", 100, "delay in miliseconds between checking for changes.")
	flag.Var(&skips, "ignore", "patterns to ignore when checking for changes, can use multiple times.")
	flag.StringVar(&command, "command", "", "command to run when changes are detected. Can be multiple shell commands or a single program.")
	flag.Usage = func() {
		log.Printf("Usage: {} -watch <path> -delay <miliseconds>")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if stat, err := os.Stat(watch_location); err != nil {
		log.Fatal(err)
		os.Exit(1)
	} else {
		path, _ := filepath.Abs(stat.Name())
		traverse(path)
	}
}

func traverse(path string) {
	log.Printf("Watching %s for changes.\n", path)
	first_run := true
	for {
		filepath.WalkDir(path, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				log.Fatal(err)
				return filepath.SkipDir
			}
			info, _ := entry.Info()
			log.Print(path + ": ")
			log.Println(info.Sys().(*syscall.Stat_t).Nlink)
			if val, exist := entries[path]; exist {
				fmt.Println(val)
			} else {
				fmt.Println("Does not exist")
			}
			return nil
		})
		if first_run {
			fmt.Println("First run complete")
		}
		first_run = false
		time.Sleep(time.Duration(watch_delay_milli) * time.Millisecond)
	}

}

func getStringFromInfo(info os.FileInfo) string {
	stat := info.Sys().(*syscall.Stat_t)
	return fmt.Sprintf("%d", stat.Nlink)
}
