package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

func isTopDir() bool {
	wd, _ := os.Getwd()
	return wd == filepath.Dir(wd)
}

func isGit() bool {
	_, err := os.Stat("./.git")
	return err == nil || os.IsExist(err)
}

func main() {
	// Walk up until we hit a git repo
	if len(os.Args) != 2 {
		fmt.Println("Please provide a path to the desired changelog file")
		os.Exit(1)
	}
	for ; !isTopDir() && !isGit(); os.Chdir("..") {
		wd, _ := os.Getwd()
		fmt.Println(wd + " is not a git repo...")
	}
	if !isGit() {
		fmt.Println("Not in a git repo!")
		os.Exit(2)
	}

	file, err := os.Open("./.git/logs/HEAD")
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
