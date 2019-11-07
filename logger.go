package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/ttacon/chalk"
)

type Logger struct {
	DebugLevel bool
	Quiet      bool
}

func (l Logger) Debug(a ...interface{}) {
	if l.DebugLevel {
		//windows < 10 doesn't supports ANSI
		if runtime.GOOS == "windows" {
			fmt.Println(a)
		} else {
			fmt.Println(chalk.Cyan, a, chalk.Reset)
		}
	}
}

func (l Logger) Info(a ...interface{}) {
	if !l.Quiet {
		if runtime.GOOS == "windows" {
			fmt.Println(a)
		} else {
			fmt.Println(chalk.Green, a, chalk.Reset)
		}
	}
}

func (l Logger) Warn(a ...interface{}) {
	if runtime.GOOS == "windows" {
		fmt.Println(a)
	} else {
		fmt.Println(chalk.Yellow, a, chalk.Reset)
	}
}

func (l Logger) Fatal(a ...interface{}) {
	if runtime.GOOS == "windows" {
		fmt.Println(a)
	} else {
		fmt.Println(chalk.Red, a, chalk.Reset)
	}
	os.Exit(1)
}
