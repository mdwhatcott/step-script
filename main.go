package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

var Version = "dev"

func main() {
	log.SetFlags(0)

	stdin := false
	flag.BoolVar(&stdin, "stdin", stdin, "When set, read from stdin first, then any files named by non-flag args.")
	flag.Usage = func() {
		log.Println("" +
			"This script accepts file paths in non-flag args which contain commands that will be executed (w/ bash -c '...'), one at a time, each after a prompt.",
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	log.Println("step-script @", Version)

	if stdin {
		commandsExecuted := executeSteps(os.Stdin)
		log.Printf("Finished running %d commands from stdin.", commandsExecuted)
	}

	paths := flag.Args()
	if !stdin && len(paths) == 0 {
		log.Fatalln("At least one path is required when not reading from stdin.")
	}

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			log.Fatalln(err)
		}
		commandsExecuted := executeSteps(file)
		_ = file.Close()
		log.Printf("Finished running %d commands in %s.", commandsExecuted, path)
	}
}

func executeSteps(reader io.ReadCloser) (result int) {
	defer func() { _ = reader.Close() }()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		fmt.Print("\n$ ", line, "    # Execute? [Y/n] ")
		if !yes() {
			log.Println("Skipping step...")
			continue
		}

		log.Println()

		shell(line)

		result++
	}
	err := scanner.Err()
	if err != nil {
		log.Fatalln("scanner err:", err)
	}
	return result
}

func shell(line string) {
	command := exec.Command("bash", "-c", line)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func yes() bool {
	response := strings.ToLower(strings.TrimSpace(prompt()))
	return response == "" || response == "y" || response == "yes"
}

func prompt() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}
