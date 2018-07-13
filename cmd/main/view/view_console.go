package view

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/mgutz/ansi"
	"github.com/milanaleksic/clici/model"
)

var greenFormat = ansi.ColorFunc("green+b+h")
var blueFormat = ansi.ColorFunc("blue+b+h")
var magentaFormat = ansi.ColorFunc("magenta+b")
var grayFormat = ansi.ColorFunc("grey")
var whiteFormat = ansi.ColorFunc("white+b+h")
var redFormat = ansi.ColorFunc("red+b")
var yellowFormat = ansi.ColorFunc("yellow+b+h")
var resetFormat = ansi.ColorCode("reset")

// ConsoleInterface is a View that dumps continuously output to console
// in each iteration, making it possible to work in case ncurses-like
// "advanced" interface can't work for some reason
type ConsoleInterface struct {
}

func registerInterruptListener(feedbackChannel chan Command) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		feedbackChannel <- CreateCmdShutdownGroup()
	}()
}

// NewConsoleInterface creates a ConsoleInterface, with a feedbackChannel to be used
// for async command feedback sending, based on keyboard commands by users
func NewConsoleInterface(feedbackChannel chan Command) (view *ConsoleInterface) {
	fmt.Println("Loading console interface...")
	view = &ConsoleInterface{}
	registerInterruptListener(feedbackChannel)
	return
}

func (ui *ConsoleInterface) friendlyCurrentStatus(buildStatus model.JobState) string {
	switch {
	case buildStatus.Building:
		return withResetOnEnd(blueFormat(buildingChar()))
	default:
		return friendlyPreviousState(&buildStatus)
	}
}

func friendlyPreviousState(buildStatus *model.JobState) string {
	switch {
	case buildStatus.PreviousState == model.Failure:
		return withResetOnEnd(redFormat(failedChar()))
	case buildStatus.PreviousState == model.Success:
		return withResetOnEnd(greenFormat(successChar()))
	case buildStatus.PreviousState == model.Undefined:
		return withResetOnEnd(magentaFormat(undefinedChar()))
	case buildStatus.PreviousState == model.Disabled:
		return withResetOnEnd(grayFormat(undefinedChar()))
	case buildStatus.PreviousState == model.Unknown:
		return withResetOnEnd(whiteFormat(unknownChar()))
	default:
		return withResetOnEnd(whiteFormat(unknownChar()))
	}
}

// PresentState comes from View and is a call that is used to ask the view
// to refresh itself based on current model state
func (ui *ConsoleInterface) PresentState(state *model.State) {
	output := "\n\n\n"
	if state.Error != nil {
		output = output + redFormat(fmt.Sprintf("Could not fetch running jobs: %v\n", state.Error)) + resetFormat
	} else {
		for i, jobState := range state.JobStates {
			output = output + string(itoidrune(i)) + " "
			if jobState.Error != nil {
				output = output + fmt.Sprintf("%30v %v%v, %v %v\n", yellowFormat(jobState.JobName), ui.friendlyCurrentStatus(jobState), redFormat(", but REST processing had an error: "), jobState.Error, resetFormat)
			} else if jobState.Building {
				output = output + fmt.Sprintf("%30v %v by %v (%v) was %v by %v\n", yellowFormat(jobState.JobName), ui.friendlyCurrentStatus(jobState), yellowFormat(jobState.CausesFriendly),
					jobState.Time, friendlyPreviousState(&jobState), jobState.CulpritsFriendly)
			} else {
				output = output + fmt.Sprintf("%30v %v by %v (%v)\n", yellowFormat(jobState.JobName), ui.friendlyCurrentStatus(jobState), yellowFormat(jobState.CausesFriendly),
					jobState.Time)
				if jobState.PreviousState == model.Success {
					output = output + fmt.Sprintf("%30v %v by %v (%v)\n", yellowFormat(jobState.JobName), ui.friendlyCurrentStatus(jobState), yellowFormat(jobState.CausesFriendly),
						jobState.Time)
				} else {
					output = output + fmt.Sprintf("%30v %v by %v (%v)\n", yellowFormat(jobState.JobName), ui.friendlyCurrentStatus(jobState), redFormat(jobState.CausesFriendly),
						jobState.Time)
				}
			}
		}
	}
	fmt.Printf("%vStatus fetched @ %v\n", output, time.Now().Format(time.RFC822))
}

// Close comes from View and is a call that is used to ask the view
// to close itself when application goes down as a result of shutdown command
func (ui *ConsoleInterface) Close() {
	// no operation
}

func withResetOnEnd(withFormatting string) string {
	return withFormatting + resetFormat
}
