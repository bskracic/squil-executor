package runtime

type Specs struct {
	Lang      string
	Image     string
	ExtraOpts string
}

type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type Runtime interface {
	// Create runtime sandbox environment for safe code execution and return its id
	Prepare(s Specs) string
	// Execute given command with optional input using specified sandbox environment
	Exec(id string, cmd []string, ch chan *ExecResult) error
	// Copy files to sandbox environment
	CopyFile(id string, content string, filename string, dst string) error
	CreateDir(id string, dirpath string) error
	// Stop the execution of a program and remove it
	Kill(id string)
}
