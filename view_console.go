package main

import (
	"time"
	"fmt"
	"os"
	"os/signal"
	"github.com/mgutz/ansi"
)

var greenFormat func(string) string = ansi.ColorFunc("green+b+h")
var blueFormat func(string) string = ansi.ColorFunc("blue+b+h")
var redFormat func(string) string = ansi.ColorFunc("red+b+h")
var yellowFormat func(string) string = ansi.ColorFunc("yellow+b+h")
var resetFormat string = ansi.ColorCode("reset")

type ConsoleInterface struct {
}

func registerInterruptListener(feedbackChannel chan string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		feedbackChannel <- "shutdown"
	}()
}

func NewConsoleInterface(feedbackChannel chan string) (view *ConsoleInterface, err error) {
	fmt.Println("Loading console interface...")
	view = &ConsoleInterface{}
	registerInterruptListener(feedbackChannel)
	return
}

func (ui *ConsoleInterface) friendlyCurrentStatus(buildStatus JobState) string {
	var withResetOnEnd = func(withFormatting string) string {
		return withFormatting + resetFormat
	}
	switch {
	case buildStatus.Building:
		return withResetOnEnd(blueFormat("⟳"))
	case buildStatus.PreviousState == Failure:
		return withResetOnEnd(redFormat("✖"))
	case buildStatus.PreviousState == Success:
		return withResetOnEnd(greenFormat("✓"))
	}
	return withResetOnEnd(redFormat("?"))
}

func (ui *ConsoleInterface) PresentState(state *State) {
	var output string = "\n\n\n"
	if state.Error != nil {
		output = output + redFormat(fmt.Sprintf("Could not fetch running jobs: %v\n", state.Error)) + resetFormat
	} else {
		for _, jobState := range state.JobStates {
			if jobState.Error != nil {
				output = output + fmt.Sprintf("%30v %v%v, %v %v\n", yellowFormat(jobState.JobName), ui.friendlyCurrentStatus(jobState), redFormat(", but REST processing had an error: "), jobState.Error, resetFormat)
			} else if jobState.Building {
				output = output + fmt.Sprintf("%30v %v by %v (%v) was %v\n", yellowFormat(jobState.JobName), ui.friendlyCurrentStatus(jobState), yellowFormat(jobState.CausesFriendly),
					jobState.Time, ui.previousStateFriendlyIfBuilding(&jobState))
			} else {
				if jobState.PreviousState == Success {
					output = output + fmt.Sprintf("%30v %v %v\n", yellowFormat(jobState.JobName), ui.friendlyCurrentStatus(jobState), yellowFormat(jobState.CausesFriendly))
				} else {
					output = output + fmt.Sprintf("%30v %v %v\n", yellowFormat(jobState.JobName), ui.friendlyCurrentStatus(jobState), redFormat(jobState.CausesFriendly))
				}
			}
		}
	}
	fmt.Printf("%vStatus fetched @ %v\n", output, time.Now().Format(time.RFC822))
}

func (ui *ConsoleInterface) previousStateFriendlyIfBuilding(state *JobState) string {
	var withResetOnEnd = func(withFormatting string) string {
		return withFormatting + resetFormat
	}
	if state.Building {
		switch {
		case state.PreviousState == Success:
			return withResetOnEnd(greenFormat("✓"))
		case state.PreviousState == Failure:
			return withResetOnEnd(redFormat("✖"))
		}
	}
	return ""
}


func (ui *ConsoleInterface) Close() {
	// no operation
}