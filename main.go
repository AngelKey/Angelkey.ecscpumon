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
	var ezkey, statName string
	flag.StringVar(&statName, "stat-name", "cpu usage", "Stat name")
	flag.StringVar(&ezkey, "stathat-ezkey", "x", "StatHat EZ Key")
	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "must supply a program name\n")
		return 1
	}
	if len(statName) == 0 {
		fmt.Fprintf(os.Stderr, "must provide a stat name\n")
	}
	procName := flag.Args()[0]
	m := newMonitor(procName, ezkey, statName)
	if err := m.monitorProcess(); err != nil {
		fmt.Fprintf(os.Stderr, "failure to monitor: %s", err)
		return 1
	}

	return 0
}

type monitor struct {
	procName, ezkey, statName string
}

func newMonitor(procName, ezkey, statName string) *monitor {
	return &monitor{
		ezkey:    ezkey,
		procName: procName,
		statName: statName,
	}
}

func (m *monitor) value(value float64) {
	stathat.DefaultReporter.PostEZValue(m.statName, m.ezkey, value)
}

func (m *monitor) monitorProcess() error {
	cmd := exec.Command("pgrep", m.procName)
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
			cputok := toks[3]
			cpu, err := strconv.ParseFloat(cputok, 64)
			if err != nil {
				continue
			}
			fmt.Printf("%d\n", int(cpu))
			m.value(cpu)
		}
	}()

	return cmd.Wait()
}
