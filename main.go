package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var categoryEntries map[string]string = make(map[string]string)

func isTopDir() bool {
	wd, _ := os.Getwd()
	return wd == filepath.Dir(wd)
}

func isGit() bool {
	_, err := os.Stat("./.git")
	return err == nil || os.IsExist(err)
}

func processCommitMessage(msg string) {
	if strings.Count(msg, ":") < 2 {
		return
	}
	categoryMessages := strings.Split(msg, ";")
	firstCategory := strings.Split(categoryMessages[0], ":")
	if len(firstCategory) != 3 {
		return
	}
	ticketName := firstCategory[0]
	fmt.Println("Found ticket:" + ticketName)
	categoryEntries[firstCategory[1]] = firstCategory[2]
	for i := 1; i < len(categoryMessages); i++ {
		entries := strings.Split(categoryMessages[i], ":")
		if len(entries) != 2 {
			continue
		}
		categoryEntries[entries[0]] = entries[1]
	}
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
		messageSplit := strings.SplitN(scanner.Text(), "commit: ", 2)
		if len(messageSplit) != 2 {
			continue
		}
		processCommitMessage(messageSplit[1])
	}
	fmt.Println(categoryEntries)
}
