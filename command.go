package main

type Command struct {
	group string
	job   int
}

const (
	CmdShutdownGroup        = "shutdown"
	CmdCloseGroup           = "close"
	CmdShowHelpGroup        = "showHelp"
	CmdOpenCurrentJobGroup  = "openCurrentJob"
	CmdOpenPreviousJobGroup = "openPreviousJob"
	CmdTestsForJobGroup     = "openTests"
)

func CmdShutdown() Command {
	return Command{group: CmdShutdownGroup}
}
func CmdClose() Command {
	return Command{group: CmdCloseGroup}
}
func CmdShowHelp() Command {
	return Command{group: CmdShowHelpGroup}
}

func CmdOpenCurrentJob() Command {
	return Command{group: CmdOpenCurrentJobGroup}
}

func CmdOpenPreviousJob() Command {
	return Command{group: CmdOpenPreviousJobGroup}
}

func CmdTestsForJob() Command {
	return Command{group: CmdTestsForJobGroup}
}
