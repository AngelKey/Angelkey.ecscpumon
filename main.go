package main

import (
	"bufio"
	"flag"
	"fmt"
	stathat "github.com/stathat/go"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	rc := mainInner()
	os.Exit(rc)
}

func mainInner() int {
	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "must supply a program name\n")
		return 1
	}
	procName := flag.Args()[0]
	if err := monitorProcess(procName); err != nil {
		fmt.Fprintf(os.Stderr, "failure to monitor: %s", err)
		return 1
	}

	return 0
}

func monitorProcess(procName string) error {
	cmd := exec.Command("pgrep", procName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	sout := strings.TrimSpace(string(out))
	if _, err := strconv.Atoi(sout); err != nil {
		return err
	}
	cmd = exec.Command("pidstat", "-p", sout, "1")
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	scanner := bufio.NewScanner(pipe)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			toks := strings.Fields(line)
			if len(toks) < 5 {
				continue
			}
			cputok := toks[4]
			cpu, err := strconv.ParseFloat(cputok, 64)
			if err != nil {
				continue
			}
			fmt.Printf("%d\n", int(cpu))
		}
	}()

	return cmd.Wait()
}
