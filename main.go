package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
	skips             ignores           = ignores([]string{})
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

	var patterns string
	for _, pattern := range skips {
		patterns += pattern + "|"
	}
	patterns = strings.TrimRight(patterns, "|")
	regex := regexp.MustCompile(patterns)
	first_run := true
	for {
		filepath.WalkDir(path, func(path string, entry os.DirEntry, err error) error {
			// check for error in walking path, skip dir if error
			if err != nil {
				log.Fatal(err)
				return filepath.SkipDir
			}
			// check if entry is to be skipped based on ignore patterns
			if regex.Match([]byte(entry.Name())) {
				if first_run {
					log.Println("Skipped " + entry.Name())
				}
				return nil
			}
			//
			//
			// Check for changes
			if val, exist := entries[path]; exist {
				// entry exists in map, check if changed
				fmt.Println(val == getStringFromInfo(entry))
			} else {
				if !first_run {
					// this means this is a new entry in watched directory
					fmt.Println("Does not exist: " + path)
				} else {
					// this is first run, so add to map
					entries[path] = getStringFromInfo(entry)
				}
			}
			return nil
		})
		// this only to be run for first run. Then set the first run false
		if first_run {
			fmt.Println("First run complete")
			first_run = false
		}
		fmt.Println(entries)
		time.Sleep(time.Duration(watch_delay_milli) * time.Millisecond)
	}

}

func getStringFromInfo(dir os.DirEntry) string {
	info, _ := dir.Info()
	stat := info.Sys().(*syscall.Stat_t)
	fmt.Sprintf("%d%d%d%d", stat.Nlink, stat.Ino, stat.Size, stat.Mtimespec.Sec)
	return fmt.Sprintf("%d%d%d%d", stat.Nlink, stat.Ino, stat.Size, stat.Mtimespec.Sec)
}
