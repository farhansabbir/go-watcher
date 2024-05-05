package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

var (
	watch_location    string = "./"
	watch_delay_milli int    = 100
	command           string = ""
)

func init() {
	flag.StringVar(&watch_location, "watch", "./", "location to watch for changes.")
	flag.IntVar(&watch_delay_milli, "delay", 100, "delay in miliseconds between checking for changes.")
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
	filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
			return filepath.SkipDir
		}
		if d.IsDir() {
			log.Printf("Traversing %s\n", path)
			traverse(path)
		}
		return nil
	})
}
