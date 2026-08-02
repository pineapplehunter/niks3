package main

import (
	"fmt"

	"github.com/mattn/go-isatty"
)

type outputMode string

const (
	outputAuto     outputMode = "auto"
	outputLog      outputMode = "log"
	outputProgress outputMode = "progress"
)

func selectOutputMode(value string, terminal bool) (outputMode, error) {
	mode := outputMode(value)
	switch mode {
	case outputAuto:
		if terminal {
			return outputProgress, nil
		}

		return outputLog, nil
	case outputLog, outputProgress:
		return mode, nil
	default:
		return "", fmt.Errorf("invalid --output %q (must be auto, log, or progress)", value)
	}
}

func stderrIsTerminal() bool {
	fd := uintptr(2)

	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
