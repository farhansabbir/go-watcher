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
	"time"

	"github.com/farhansabbir/go-fswatcher/lib"
)

type ignores []string

var (
	watch_location    string = "./"
	watch_delay_milli int    = 100
	// command           string            = ""
	entries map[string]string = make(map[string]string)
	skips   ignores           = ignores([]string{})
	Version                   = "0.0.1"
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
	// flag.StringVar(&command, "command", "", "Parameterless command/script to run when changes are detected. Pass platform specific script for commands with arguments or multiple commands.")
	flag.Usage = func() {
		fmt.Printf(`Usage: %s -watch <path> -delay <miliseconds>`, flag.CommandLine.Name())
		fmt.Println()
		fmt.Println("Version: " + Version)
		flag.PrintDefaults()
	}
}

func firstRun(rootpath string) {
	var patterns string
	for _, pattern := range skips {
		patterns += pattern + "|"
	}
	patterns = strings.TrimRight(patterns, "|")
	regex := regexp.MustCompile(patterns)
	filepath.WalkDir(rootpath, func(path string, entry os.DirEntry, err error) error {
		// check for error in walking path, skip dir if error
		if err != nil {
			if os.IsPermission(err) {
				log.Fatal(err)
			}
			return filepath.SkipDir
		}

		// check if entry is to be skipped based on ignore patterns
		if regex.Match([]byte(entry.Name())) && regex.String() != "" {
			return filepath.SkipDir
		}
		entries[path] = lib.GetStringFromInfo(entry)
		return nil
	})
	time.Sleep(time.Duration(watch_delay_milli) * time.Millisecond)
	// log.Println("First run setup complete for " + rootpath)
}

func main() {
	flag.Parse()
	if stat, err := os.Stat(watch_location); err != nil {
		log.Fatal(err)
		os.Exit(1)
	} else {
		path, _ := filepath.Abs(stat.Name())
		firstRun(path)
		watch(path)
	}
}

func watch(rootpath string) {
	log.Printf("Watching %s for changes.\n", rootpath)

	var patterns string
	for _, pattern := range skips {
		patterns += pattern + "|"
	}
	patterns = strings.TrimRight(patterns, "|")
	regex := regexp.MustCompile(patterns)
	for {
		filepath.WalkDir(rootpath, func(path string, entry os.DirEntry, err error) error {
			// check for error in walking path, skip dir if error
			if err != nil {
				log.Fatal(err)
				return filepath.SkipDir
			}

			// check if entry is to be skipped based on ignore patterns
			if regex.Match([]byte(entry.Name())) && regex.String() != "" {
				return filepath.SkipDir
			}
			if val, exist := entries[path]; exist {
				// entry exists in map, check if changed
				if val != lib.GetStringFromInfo(entry) {
					log.Println(path + " changed")
					entries[path] = lib.GetStringFromInfo(entry)
				}
			} else {
				entries[path] = lib.GetStringFromInfo(entry)
			}
			return nil
		})
		time.Sleep(time.Duration(watch_delay_milli) * time.Millisecond)
	}

}
