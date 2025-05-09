package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/farhansabbir/go-fswatcher/lib"
)

type ignores []string

var (
	watch_location    string            = "./"
	watch_delay_milli int               = 100
	command           string            = ""
	entries           map[string]string = make(map[string]string)
	skips             ignores           = ignores([]string{})
	Version                             = "0.0.1"
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
	flag.StringVar(&command, "command", "", "Parameterless command/script to run when changes are detected. Pass platform specific script for commands with arguments or multiple commands.")
	flag.Usage = func() {
		fmt.Printf(`Usage: %s -watch <path> -delay <miliseconds>`, flag.CommandLine.Name())
		fmt.Println()
		fmt.Println("Version: " + Version)
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

func traverse(rootpath string) {
	log.Printf("Watching %s for changes.\n", rootpath)

	var patterns string
	for _, pattern := range skips {
		patterns += pattern + "|"
	}
	patterns = strings.TrimRight(patterns, "|")
	regex := regexp.MustCompile(patterns)
	first_run := true
	for {
		filepath.WalkDir(rootpath, func(path string, entry os.DirEntry, err error) error {
			// check for error in walking path, skip dir if error
			if err != nil {
				log.Fatal(err)
				return filepath.SkipDir
			}

			// check if entry is to be skipped based on ignore patterns
			if regex.Match([]byte(entry.Name())) && regex.String() != "" {
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
				if val != lib.GetStringFromInfo(entry) {
					if command != "" {
						log.Println("Change: '" + path + "'")
						var procAttr os.ProcAttr
						var output strings.Builder
						procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
						// commandsplit := strings.Split(flag.Arg(0), " ")
						ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
						defer cancel()
						proc := exec.CommandContext(ctx, command)
						proc.Stdout = &output

						if err = proc.Run(); err != nil {
							log.Println(err)
							log.Println("Command '" + command + "' did not run successfully.")
						} else {
							fmt.Println(output.String())
						}

						// proc, _ := os.StartProcess(command, []string{command}, &procAttr)
						// state, err := proc.Wait()
						// if err != nil {
						// 	log.Println(err.Error())
						// }
						// log.Println(state.Success())
					} else {
						log.Println("Change: '" + path + "'")
					}
					entries[path] = lib.GetStringFromInfo(entry)
				}
			} else {
				if !first_run {
					log.Println("New: '" + path + "'")
				}

				entries[path] = lib.GetStringFromInfo(entry)
			}
			return nil
		})
		// this only to be run for first run. Then set the first run false
		if first_run {
			first_run = false
		}
		time.Sleep(time.Duration(watch_delay_milli) * time.Millisecond)
	}

}
