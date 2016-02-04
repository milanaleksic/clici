package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
	"strconv"
	"time"
)

type CUIInterface struct {
	gui             *gocui.Gui
	statusField     *gocui.View
	feedbackChannel chan Command
	tableStart      int
}

func friendlyKnownStatus(buildStatus JobState) string {
	switch {
	case buildStatus.PreviousState == Failure:
		return failedChar()
	case buildStatus.PreviousState == Success:
		return successChar()
	}
	return ""
}

func (ui *CUIInterface) PresentState(state *State) {
	if state.Error != nil || len(state.JobStates) == 0 {
		ui.errorDialog(state)
		ui.bottomLine()
		return
	}
	if len(state.FailedTests) != 0 {
		ui.informationDialogOfTests(state)
		ui.bottomLine()
		return
	}
	ui.gui.SetLayout(func(gui *gocui.Gui) error {
		lengthForJobNames := state.MaxLengthOfName()
		if v, err := gui.SetView("job_id", 0, ui.tableStart, 3, 2*len(state.JobStates)+4); err != nil {
			check_cui(err)
			v.Frame = false
			v.FgColor = gocui.ColorWhite
			for i, _ := range state.JobStates {
				fmt.Fprintf(v, "%s\n", string(itoidrune(i)))
			}
		}
		if v, err := gui.SetView("job_name", 2, ui.tableStart, lengthForJobNames+3, 2*len(state.JobStates)+4); err != nil {
			check_cui(err)
			v.Frame = false
			v.FgColor = gocui.ColorYellow
			for _, jobState := range state.JobStates {
				fmt.Fprintf(v, "%"+strconv.Itoa(lengthForJobNames)+"v\n", jobState.JobName)
			}
		}
		for index, jobState := range state.JobStates {
			ui.showAJobDetailColumns(&jobState, index, lengthForJobNames)
		}
		ui.topLine(lengthForJobNames)
		ui.bottomLine()
		if state.ShowHelp {
			ui.helpDialog()
		}
		return nil
	})
}

func (ui *CUIInterface) showAJobDetailColumns(jobState *JobState, index int, lengthForJobNames int) {
	gui := ui.gui
	maxX, _ := gui.Size()
	tableStart := ui.tableStart
	if v, err := gui.SetView(fmt.Sprintf("building_job_%v", index), lengthForJobNames+3, tableStart+index, lengthForJobNames+5, tableStart+index+2); err != nil {
		check_cui(err)
		v.Frame = false
		v.FgColor = gocui.ColorBlue | gocui.AttrBold
		if jobState.Building {
			switch {
			case jobState.Building:
				fmt.Fprint(v, buildingChar())
			}
		}
	}
	if v, err := gui.SetView(fmt.Sprintf("curr_job_status_%v", index), lengthForJobNames+5, tableStart+index, lengthForJobNames+7, tableStart+index+2); err != nil {
		check_cui(err)
		v.Frame = false
		switch {
		case jobState.PreviousState == Failure:
			v.FgColor = gocui.ColorRed | gocui.AttrBold
		case jobState.PreviousState == Success:
			v.FgColor = gocui.ColorGreen | gocui.AttrBold
		default:
			v.FgColor = gocui.ColorWhite | gocui.AttrBold
		}
		fmt.Fprintf(v, "%v", friendlyKnownStatus(*jobState))
	}
	if v, err := gui.SetView(fmt.Sprintf("curr_job_description_%v", index), lengthForJobNames+7, tableStart+index, maxX, tableStart+index+2); err != nil {
		check_cui(err)
		v.Frame = false
		if jobState.Error != nil {
			v.FgColor = gocui.ColorRed | gocui.AttrBold
			fmt.Fprintf(v, "API processing had an error: %v", jobState.Error)
		} else if jobState.Building {
			if jobState.PreviousState == Success {
				v.FgColor = gocui.ColorBlue
				fmt.Fprintf(v, "by %v (%v)", jobState.CausesFriendly, jobState.Time)
			} else {
				v.FgColor = gocui.ColorRed | gocui.AttrBold
				fmt.Fprintf(v, "by %v (%v); failed by %v", jobState.CausesFriendly, jobState.Time, jobState.CulpritsFriendly)
			}
		} else {
			if jobState.PreviousState == Success {
				v.FgColor = gocui.ColorGreen
				fmt.Fprintf(v, "%v", jobState.CausesFriendly)
			} else {
				v.FgColor = gocui.ColorRed | gocui.AttrBold
				fmt.Fprintf(v, "%v; failed by %v", jobState.CausesFriendly, jobState.CulpritsFriendly)
			}
		}
	}
}

func NewCUIInterface(feedbackChannel chan Command) (view *CUIInterface, err error) {
	view = &CUIInterface{
		gui:             gocui.NewGui(),
		feedbackChannel: feedbackChannel,
	}
	if err = view.gui.Init(); err != nil {
		return
	}
	view.gui.BgColor = gocui.ColorDefault
	view.gui.FgColor = gocui.ColorWhite
	view.gui.SetLayout(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView("center", maxX/2-24, maxY/2-2, maxX/2+23, maxY/2+1); err != nil {
			check_cui(err)
			v.Frame = false
			fmt.Fprintln(v, " Jenkins Ping\n https://github.com/milanaleksic/jenkins_ping")
		}
		return nil
	})
	quit := func(g *gocui.Gui, v *gocui.View) error {
		return gocui.Quit
	}
	if err = view.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return
	}
	if err = view.gui.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return
	}
	if err = view.gui.SetKeybinding("", '?', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		view.feedbackChannel <- CmdShowHelp()
		return nil
	}); err != nil {
		return
	}
	var cmd = CmdOpenCurrentJob()
	setCommand := func(x Command) func(*gocui.Gui, *gocui.View) error {
		return func(g *gocui.Gui, v *gocui.View) error {
			cmd = x
			return nil
		}
	}
	if err = view.gui.SetKeybinding("", 'p', gocui.ModNone, setCommand(CmdOpenPreviousJob())); err != nil {
		return
	}
	if err = view.gui.SetKeybinding("", 't', gocui.ModNone, setCommand(CmdTestsForJob())); err != nil {
		return
	}
	for i := 0; i < 20; i++ {
		var localizedI = i
		if err = view.gui.SetKeybinding("", itoidrune(i), gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			cmd.job = localizedI
			view.feedbackChannel <- cmd
			cmd = CmdOpenCurrentJob()
			return nil
		}); err != nil {
			return
		}
	}
	if err = view.gui.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		view.feedbackChannel <- CmdClose()
		return nil
	}); err != nil {
		return
	}
	go func() {
		err = view.gui.MainLoop()
		if err != nil && err != gocui.Quit {
			log.Panicln(err)
		}
		view.gui.Close()
		feedbackChannel <- CmdShutdown()
	}()
	return
}

func (ui *CUIInterface) Close() {
	ui.gui.Close()
}

func (ui *CUIInterface) errorDialog(state *State) {
	ui.gui.SetLayout(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView("center", 1, maxY/2-1, maxX-1, maxY/2+2); err != nil {
			check_cui(err)
			v.FgColor = gocui.ColorRed
			fmt.Fprintln(v, fmt.Sprintf("Error: %v\n", state.Error))
		}
		return nil
	})
}

func (ui *CUIInterface) bottomLine() (err error) {
	maxX, maxY := ui.gui.Size()
	fetchedMessage := fmt.Sprintf(" @ %v ", time.Now().Format(time.RFC822))
	if v, err := ui.gui.SetView("bottom_left", -1, maxY-2, maxX-len(fetchedMessage)+1, maxY); err != nil {
		check_cui(err)
		v.BgColor = gocui.ColorBlack
		v.FgColor = gocui.ColorWhite
		v.Frame = false
		fmt.Fprintf(v, "<q>: Quit   <id>: Go to job   <?>: Show all commands")
	}
	if v, err := ui.gui.SetView("bottom_right", maxX-len(fetchedMessage), maxY-2, maxX, maxY); err != nil {
		check_cui(err)
		v.BgColor = gocui.ColorBlack
		v.FgColor = gocui.ColorWhite
		v.Frame = false
		fmt.Fprintf(v, fetchedMessage)
	}
	return
}

func (ui *CUIInterface) topLine(lengthForJobNames int) (err error) {
	maxX, _ := ui.gui.Size()
	if v, err := ui.gui.SetView("top", -1, -1, maxX, 1); err != nil {
		if err != gocui.ErrorUnkView {
			return err
		}
		v.BgColor = gocui.ColorDefault
		v.FgColor = gocui.ColorWhite
		v.Frame = false
		fmt.Fprintf(v, "ID %"+strconv.Itoa(lengthForJobNames)+"v B S DESCRIPTION", "NAME")
	}
	return
}

func (ui *CUIInterface) helpDialog() {
	ui.gui.SetLayout(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView("center", maxX/2-26, maxY/2-3, maxX/2+26, maxY/2+3); err != nil {
			check_cui(err)
			v.FgColor = gocui.ColorWhite
			v.Overwrite = false
			fmt.Fprintln(v, `
			              q - Quit
			           <id> - Open Last Job URL
			         p+<id> - Open Last Completed Job URL
			         t+<id> - Show Test failures (NYI)
			          Enter - Close Help
			`)
		}
		return nil
	})
}

func (ui *CUIInterface) informationDialogOfTests(state *State) {
	ui.gui.SetLayout(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView("center", 2, 1, maxX-3, maxY-3); err != nil {
			check_cui(err)
			maxLength := maxX - 6
			var output string = fmt.Sprintf("Failed tests (%d of them): \n", len(state.FailedTests))
			for _, failedTest := range state.FailedTests {
				if len(failedTest) > maxLength {
					output = output + fmt.Sprintf("%s\n", failedTest[len(failedTest)-maxLength:])
				} else {
					output = output + fmt.Sprintf("%s\n", failedTest)
				}
			}
			v.FgColor = gocui.ColorWhite
			v.Overwrite = false
			fmt.Fprintln(v, output)
		}
		return nil
	})
}
