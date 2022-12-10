package main

import (
	"fmt"
	"grep/workers"
	"grep/worklist"
	"os"
	"sync"

	"github.com/alexflint/go-arg"
)

var args struct {
	Path      string `arg:"positional, required" help:"Path to the file"`
	Pattern   string `arg:"positional, required"  help:"Pattern to look for in file"`
	Workers   int    `arg:"-t,--threads" default:"1" help:"Number of threads to use if the search is recursive"`
	Recursive bool   `arg:"-R,--recursive" default:"false" help:"Recursive search in a directory and will automatically use 8 threads"`
}

func main() {
	cli := arg.MustParse(&args)
	if !args.Recursive {
		info, err := os.Stat(args.Path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if !info.Mode().IsDir() {
			workers.FindInFile(args.Path, args.Pattern)
			return
		} else {
			cli.Fail("-R can only be used with directories")
			os.Exit(1)
		}
	}
	args.Workers = 8
	wl := worklist.NewWorklist(100)
	results := worklist.NewResults(100)
	var workersWg sync.WaitGroup

	workersWg.Add(1)
	go func() {
		defer workersWg.Done()
		workers.DiscoverDirs(args.Path, &wl)
		//* This is very important because once a channel is closed, it sends default values
		//* which means a "" will be sent on the end of the channel, ad infinitum.
		//* Once all paths were consumed from the channel, these "" will be sent and once
		//* workers receive them they'll terminate. If this doesn't happen, the program
		//* will deadlock.
		close(wl)
	}()

	for i := 0; i < args.Workers; i++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()
			for {
				path := wl.Get()
				if path == "" {
					return
				}
				workers.FindInFile(path, args.Pattern)
			}
		}()
	}
	var blockWg sync.WaitGroup
	blockChannel := make(chan struct{})

	blockWg.Add(1)
	go func() {
		defer blockWg.Done()
		workersWg.Wait()
		close(blockChannel)
	}()

finalLoop:
	for {
		select {
		case result := <-results:
			fmt.Printf("%q[%d]: %q\n", result.Path, result.LineNum, result.Line)
		case <-blockChannel:
			if len(results) == 0 {
				break finalLoop
			}
		}
	}
	blockWg.Wait()
}
