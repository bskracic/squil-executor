package runner

type RunStatus string

const (
	Interrupted RunStatus = "Interrupted"
	Failed                = "Failed"
	Finished              = "Finished"
)

type RunOptions struct {
	Stdin      string
	InputFiles *[]string
}

type RunResult struct {
	Result   string
	ExitCode int
	Status   RunStatus
	Compiles bool
}

type RunCtx struct {
	ContId    string
	CompileId string
}

type Runner interface {
	// Creates new runtime context
	newContext() *RunCtx
	// Compiles given source code
	compile(ctx *RunCtx, src []byte) (*RunResult, error)
	// Runs given source code, if a language is compiled, compile step is required beforehand
	exec(ctx *RunCtx, options *RunOptions) (*RunResult, error)
	// Clean up environment
	cleanUp(ctx *RunCtx)
	// Template method
	Run(src string, options *RunOptions) *RunResult

	CreateContainer() string
}
