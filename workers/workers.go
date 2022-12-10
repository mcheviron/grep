package workers

import (
	"bufio"
	"fmt"
	"grep/worklist"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func FindInFile(path, pattern string) {
	fHandle, err := os.Open(path)
	if err != nil {
		// fmt.Fprintln(os.Stderr, err)
		return
	}
	info, err := os.Stat(fHandle.Name())
	if err != nil {
		// fmt.Fprintln(os.Stderr, err)
		return
	}
	if IsExec(info.Mode()) {
		// fmt.Fprintln(os.Stderr, "Can't read an executable file")
		return
	}
	scanner := bufio.NewScanner(fHandle)
	lineNum := 1
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			fmt.Printf("%q[%d]: %q\n", path, lineNum, line)
		}
		lineNum += 1
	}

}

func IsExec(mode fs.FileMode) bool {
	return mode&0111 == 0111
}

func DiscoverDirs(path string, wl *worklist.Worklist) {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	//* The provided path is a file and not a directory
	if info.Mode().IsRegular() {
		wl.Add(path)
		return
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			DiscoverDirs(filepath.Join(path, entry.Name()), wl)
		} else if !(entry.Type()&0111 == 0111) {
			wl.Add(filepath.Join(path, entry.Name()))
		}
	}
}
