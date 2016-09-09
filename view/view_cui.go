package view

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/milanaleksic/clici/model"
)

// CUIInterface is a View that uses a ncurses-like advanced interface that
// gives a similar-to-desktop look & feed
type CUIInterface struct {
	gui             *gocui.Gui
	feedbackChannel chan Command
	tableStart      int
}

func checkCui(err error) {
	if err != gocui.ErrUnknownView {
		log.Panicf("Unexpected error occured: %v", err)
	}
}

func friendlyKnownStatus(buildStatus model.JobState) string {
	switch {
	case buildStatus.PreviousState == model.Failure:
		return failedChar()
	case buildStatus.PreviousState == model.Success:
		return successChar()
	case buildStatus.PreviousState == model.Undefined:
		return undefinedChar()
	}
	return ""
}

// PresentState comes from View and is a call that is used to ask the view
// to refresh itself based on current model state
func (ui *CUIInterface) PresentState(state *model.State) {
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
		lengthForJobNames := ui.maxLengthOfName(state)
		if v, err := gui.SetView("job_id", 0, ui.tableStart, 3, 2*len(state.JobStates)+4); err != nil {
			checkCui(err)
			v.Frame = false
			v.FgColor = gocui.ColorWhite
			for i := range state.JobStates {
				fmt.Fprintf(v, "%s\n", string(itoidrune(i)))
			}
		}
		if v, err := gui.SetView("job_name", 2, ui.tableStart, lengthForJobNames+3, 2*len(state.JobStates)+4); err != nil {
			checkCui(err)
			v.Frame = false
			v.FgColor = gocui.ColorYellow
			for _, jobState := range state.JobStates {
				fmt.Fprintf(v, "%"+strconv.Itoa(lengthForJobNames)+"v\n", jobState.JobName)
			}
		}
		for index, jobState := range state.JobStates {
			ui.showJobColumns(&jobState, index, lengthForJobNames)
		}
		ui.topLine(lengthForJobNames)
		ui.bottomLine()
		if state.ShowHelp {
			ui.helpDialog()
		}
		return nil
	})
}

func (ui *CUIInterface) showJobColumns(jobState *model.JobState, index int, lengthForJobNames int) {
	ui.showBuildFlagColumn(jobState, index, lengthForJobNames)
	ui.showJobStatusColumn(jobState, index, lengthForJobNames)
	ui.showJobDescriptionColumn(jobState, index, lengthForJobNames)
}

func (ui *CUIInterface) showBuildFlagColumn(jobState *model.JobState, index int, lengthForJobNames int) {
	if v, err := ui.gui.SetView(fmt.Sprintf("building_job_%v", index), lengthForJobNames+3, ui.tableStart+index, lengthForJobNames+5, ui.tableStart+index+2); err != nil {
		checkCui(err)
		v.Frame = false
		v.FgColor = gocui.ColorBlue | gocui.AttrBold
		if jobState.Building {
			switch {
			case jobState.Building:
				fmt.Fprint(v, buildingChar())
			}
		}
	}
}

func (ui *CUIInterface) showJobStatusColumn(jobState *model.JobState, index int, lengthForJobNames int) {
	if v, err := ui.gui.SetView(fmt.Sprintf("curr_job_status_%v", index), lengthForJobNames+5, ui.tableStart+index, lengthForJobNames+7, ui.tableStart+index+2); err != nil {
		checkCui(err)
		v.Frame = false
		switch {
		case jobState.PreviousState == model.Failure:
			v.FgColor = gocui.ColorRed | gocui.AttrBold
		case jobState.PreviousState == model.Success:
			v.FgColor = gocui.ColorGreen | gocui.AttrBold
		case jobState.PreviousState == model.Undefined:
			v.FgColor = gocui.ColorMagenta | gocui.AttrBold
		case jobState.PreviousState == model.Unknown:
			v.FgColor = gocui.ColorWhite | gocui.AttrBold
		default:
			v.FgColor = gocui.ColorWhite | gocui.AttrBold
		}
		fmt.Fprintf(v, "%v", friendlyKnownStatus(*jobState))
	}
}

func (ui *CUIInterface) showJobDescriptionColumn(jobState *model.JobState, index int, lengthForJobNames int) {
	maxX, _ := ui.gui.Size()
	if v, err := ui.gui.SetView(fmt.Sprintf("curr_job_description_%v", index), lengthForJobNames+7, ui.tableStart+index, maxX, ui.tableStart+index+2); err != nil {
		checkCui(err)
		v.Frame = false
		if jobState.Error != nil {
			v.FgColor = gocui.ColorRed | gocui.AttrBold
			fmt.Fprintf(v, "API processing had an error: %v", jobState.Error)
		} else {
			if jobState.PreviousState == model.Success {
				v.FgColor = gocui.ColorGreen
				fmt.Fprintf(v, "%v (%v)", jobState.CausesFriendly, jobState.Time)
			} else {
				v.FgColor = gocui.ColorRed | gocui.AttrBold
				fmt.Fprintf(v, "%v (%v); failed by %v", jobState.CausesFriendly, jobState.Time, jobState.CulpritsFriendly)
			}
		}
	}
}

// NewCUIInterface creates a CUIInterface (advanced console interface), with a feedbackChannel to be used
// for async command feedback sending, based on keyboard commands by users
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
		if v, err2 := g.SetView("center", maxX/2-24, maxY/2-2, maxX/2+23, maxY/2+1); err2 != nil {
			checkCui(err2)
			v.Frame = false
			fmt.Fprintln(v, " Jenkins Ping\n https://github.com/milanaleksic/clici")
		}
		return nil
	})
	view.setKeyBindings()
	go func() {
		err = view.gui.MainLoop()
		if err != nil && err != gocui.ErrQuit {
			log.Panicln(err)
		}
		view.gui.Close()
		feedbackChannel <- CmdShutdown()
	}()
	return
}

// Close comes from View and is a call that is used to ask the view
// to close itself when application goes down as a result of shutdown command
func (ui *CUIInterface) Close() {
	ui.gui.Close()
}

func (ui *CUIInterface) setKeyBindings() {
	quit := func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}
	if err := ui.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return
	}
	if err := ui.gui.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return
	}
	if err := ui.gui.SetKeybinding("", '?', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		ui.feedbackChannel <- CmdShowHelp()
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
	if err := ui.gui.SetKeybinding("", 'p', gocui.ModNone, setCommand(CmdOpenPreviousJob())); err != nil {
		return
	}
	if err := ui.gui.SetKeybinding("", 't', gocui.ModNone, setCommand(CmdTestsForJob())); err != nil {
		return
	}
	for i := 0; i < 20; i++ {
		var localizedI = i
		if err := ui.gui.SetKeybinding("", itoidrune(i), gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			cmd.Job = localizedI
			ui.feedbackChannel <- cmd
			cmd = CmdOpenCurrentJob()
			return nil
		}); err != nil {
			return
		}
	}
	if err := ui.gui.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		ui.feedbackChannel <- CmdClose()
		return nil
	}); err != nil {
		return
	}
}

func (ui *CUIInterface) errorDialog(state *model.State) {
	ui.gui.SetLayout(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView("center", 1, maxY/2-1, maxX-1, maxY/2+2); err != nil {
			checkCui(err)
			v.FgColor = gocui.ColorRed
			fmt.Fprintln(v, fmt.Sprintf("Error: %v\n", state.Error))
		}
		return nil
	})
}

func (ui *CUIInterface) bottomLine() {
	maxX, maxY := ui.gui.Size()
	fetchedMessage := fmt.Sprintf(" @ %v ", time.Now().Format(time.RFC822))
	if v, err := ui.gui.SetView("bottom_left", -1, maxY-2, maxX-len(fetchedMessage)+1, maxY); err != nil {
		checkCui(err)
		v.BgColor = gocui.ColorBlack
		v.FgColor = gocui.ColorWhite
		v.Frame = false
		fmt.Fprint(v, "<q>: Quit   <id>: Go to job   <?>: Show all commands")
	}
	if v, err := ui.gui.SetView("bottom_right", maxX-len(fetchedMessage), maxY-2, maxX, maxY); err != nil {
		checkCui(err)
		v.BgColor = gocui.ColorBlack
		v.FgColor = gocui.ColorWhite
		v.Frame = false
		fmt.Fprintf(v, fetchedMessage)
	}
	return
}

func (ui *CUIInterface) topLine(lengthForJobNames int) {
	maxX, _ := ui.gui.Size()
	if v, err := ui.gui.SetView("top", -1, -1, maxX, 1); err != nil {
		checkCui(err)
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
			checkCui(err)
			v.FgColor = gocui.ColorWhite
			v.Overwrite = false
			fmt.Fprintln(v, ""+
				"              q - Quit\n"+
				"           <id> - Open Last Job URL\n"+
				"         p+<id> - Open Last Completed Job URL\n"+
				"         t+<id> - Show Test failures\n"+
				"          Enter - Close Help\n")
		}
		return nil
	})
}

func (ui *CUIInterface) informationDialogOfTests(state *model.State) {
	ui.gui.SetLayout(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		if v, err := g.SetView("center", 2, 1, maxX-3, maxY-3); err != nil {
			checkCui(err)
			maxLength := maxX - 6
			output := fmt.Sprintf("Failed tests (%d of them): \n", len(state.FailedTests))
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

func (ui *CUIInterface) maxLengthOfName(state *model.State) (lengthForJobNames int) {
	lengthForJobNames = 10
	for _, jobState := range state.JobStates {
		if len(jobState.JobName) > lengthForJobNames {
			lengthForJobNames = len(jobState.JobName)
		}
	}
	return
}
