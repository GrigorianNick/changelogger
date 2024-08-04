package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var preamble string = "# Changelog\nAll notable changes to this project will be documented in this file.\n\nThe format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),\nand this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n"

type version struct {
	major, minor, micro int
}

func newVersion(major, minor, micro string) *version {
	v := &version{}
	v.major, _ = strconv.Atoi(major)
	v.minor, _ = strconv.Atoi(minor)
	v.micro, _ = strconv.Atoi(micro)
	return v
}

func (v *version) getString() string {
	return "[" + strconv.Itoa(v.major) + "." + strconv.Itoa(v.minor) + "." + strconv.Itoa(v.micro) + "]"
}

var categoryEntries map[string][]string = make(map[string][]string)

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
	categoryEntries[firstCategory[1]] = append(categoryEntries[firstCategory[1]], ticketName+firstCategory[2])
	for i := 1; i < len(categoryMessages); i++ {
		entries := strings.Split(categoryMessages[i], ":")
		if len(entries) != 2 {
			continue
		}
		categoryEntries[entries[0]] = append(categoryEntries[entries[0]], ticketName+entries[1])
	}
}

func findLastReleaseTime(logPath string) (time.Time, *version, error) {
	file, err := os.Open(logPath)
	if err != nil {
		return time.Time{}, &version{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`## \[(\d+)\.(\d+)\.(\d+)\]\ - (\d{4}-\d{2}-\d{2})`)
	for scanner.Scan() {
		finds := re.FindStringSubmatch(scanner.Text())
		fmt.Println(finds)
		if len(finds) == 5 {
			t, err := time.Parse("2006-01-02", finds[4])
			if err == nil {
				return t, newVersion(finds[1], finds[2], finds[3]), err
			}
		}
	}
	return time.Time{}, &version{}, nil
}

func writeEntries(filePath string, v *version) {
	file, _ := os.Create(filePath)
	defer file.Close()
	io.WriteString(file, preamble)
	n := time.Now()
	io.WriteString(file, `## `+v.getString())
	io.WriteString(file, ` - `+n.Format("2006-01-02")+"\n")
	for category, entries := range categoryEntries {
		io.WriteString(file, "\n")
		io.WriteString(file, `### `+category+"\n\n")
		for _, entry := range entries {
			io.WriteString(file, "- "+entry+"\n")
		}
	}
}

func mergeChangelogs(newPath, oldPath string) {
	newFile, _ := os.OpenFile("./TestChangelog2.md", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer newFile.Close()
	oldFile, _ := os.Open(os.Args[1])
	defer oldFile.Close()
	oldFile.Seek(int64(len(preamble)), 0)
	fmt.Println(io.Copy(newFile, oldFile))
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

	lastReleaseTime, v, err := findLastReleaseTime(os.Args[1])
	if err != nil {
		os.Exit(4)
	}

	file, err := os.Open("./.git/logs/HEAD")
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	defer file.Close()

	re := regexp.MustCompile(`\d{10}`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		messageSplit := strings.SplitN(scanner.Text(), "commit: ", 2)
		if len(messageSplit) != 2 {
			continue
		}
		val := re.FindStringSubmatch(messageSplit[0])
		i, err := strconv.ParseInt(val[0], 10, 64)
		if err != nil {
			fmt.Println(err)
		}
		t := time.Unix(i, 0)
		if t.Before(lastReleaseTime) {
			fmt.Println("Skipping")
			continue
		} else {
			fmt.Println(t)
			fmt.Println(lastReleaseTime)
			fmt.Println("---")
		}
		processCommitMessage(messageSplit[1])
	}
	fmt.Println(categoryEntries)
	v.micro++
	writeEntries("./TestChangelog2.md", v)
	mergeChangelogs("./TestChangelog2.md", os.Args[1])
}
