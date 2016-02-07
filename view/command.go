package view

// Command represents an interaction from user interface towards the dispatcher.
// Controller know how to
type Command struct {
	Group string
	Job   int
}

const (
	// CmdShutdownGroup declares an application shutdown command group. Takes no job parameter
	CmdShutdownGroup = "shutdown"
	// CmdCloseGroup declares a dialog close command group. Takes no job parameter
	CmdCloseGroup = "close"
	// CmdShowHelpGroup declares a "show the help dialog" command group. Takes no job parameter
	CmdShowHelpGroup = "showHelp"
	// CmdOpenCurrentJobGroup declares a command group to open the current running job behind a certain id
	CmdOpenCurrentJobGroup = "openCurrentJob"
	// CmdOpenPreviousJobGroup declares a command group to open the previous job behind a certain id
	CmdOpenPreviousJobGroup = "openPreviousJob"
	// CmdTestsForJobGroup declares a command group to open the dialog with failing tests in a job behind a certain id
	CmdTestsForJobGroup = "openTests"
)

// CmdShutdown creates a new command of group CmdShutdownGroup
func CmdShutdown() Command {
	return Command{Group: CmdShutdownGroup}
}

// CmdClose creates a new command of group CmdShowHelpGroup
func CmdClose() Command {
	return Command{Group: CmdShowHelpGroup}
}

// CmdShowHelp creates a new command of group CmdShowHelpGroup
func CmdShowHelp() Command {
	return Command{Group: CmdShowHelpGroup}
}

// CmdOpenCurrentJob creates a new command of group CmdOpenCurrentJobGroup
func CmdOpenCurrentJob() Command {
	return Command{Group: CmdOpenCurrentJobGroup}
}

// CmdOpenPreviousJob creates a new command of group CmdOpenPreviousJobGroup
func CmdOpenPreviousJob() Command {
	return Command{Group: CmdOpenPreviousJobGroup}
}

// CmdTestsForJob creates a new command of group CmdTestsForJobGroup
func CmdTestsForJob() Command {
	return Command{Group: CmdTestsForJobGroup}
}
