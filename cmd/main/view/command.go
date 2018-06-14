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
	// CmdRunJob runs a job with a certain ID
	CmdRunJob = "runJob"
)

// CreateCmdShutdownGroup creates a new command of group CmdShutdownGroup
func CreateCmdShutdownGroup() Command {
	return Command{Group: CmdShutdownGroup}
}

// CreateCmdCloseGroup creates a new command of group CmdShowHelpGroup
func CreateCmdCloseGroup() Command {
	return Command{Group: CmdCloseGroup}
}

// CreateCmdShowHelpGroup creates a new command of group CmdShowHelpGroup
func CreateCmdShowHelpGroup() Command {
	return Command{Group: CmdShowHelpGroup}
}

// CreateCmdOpenCurrentJobGroup creates a new command of group CmdOpenCurrentJobGroup
func CreateCmdOpenCurrentJobGroup() Command {
	return Command{Group: CmdOpenCurrentJobGroup}
}

// CreateCmdOpenPreviousJob creates a new command of group CmdOpenPreviousJobGroup
func CreateCmdOpenPreviousJob() Command {
	return Command{Group: CmdOpenPreviousJobGroup}
}

// CreateCmdTestsForJobGroup creates a new command of group CmdTestsForJobGroup
func CreateCmdTestsForJobGroup() Command {
	return Command{Group: CmdTestsForJobGroup}
}

// CreateCmdRunJob creates a new command of group CmdRunJob
func CreateCmdRunJob() Command {
	return Command{Group: CmdRunJob}
}