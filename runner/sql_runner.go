package runner

import (
	"fmt"
	"time"

	"github.com/bskracic/squil-executor/runtime"
)

const (
	imageName          = "mcr.microsoft.com/mssql/server:2022-latest"
	language           = "sql"
	contBasePath       = "/"
	executionTimeLimit = 10000
)

var executeCmd = []string{"/opt/mssql-tools/bin/sqlcmd", "-S", "sql1", "-U", "SA", "-P", "Supersecretpass123", "-s,", "-W", "-Q"}

var spec = runtime.Specs{
	Lang:  language,
	Image: imageName,
}

type SqlRunner struct {
	runtime runtime.Runtime
}

func NewSqlRunner(r runtime.Runtime) *SqlRunner {
	return &SqlRunner{runtime: r}
}

func (sr *SqlRunner) newContext() *RunCtx {
	contId := sr.runtime.Prepare(spec)
	return &RunCtx{
		ContId: contId,
	}
}

func (cpr *SqlRunner) compile(ctx *RunCtx, src string) (*RunResult, error) {
	// no compile step for this shit
	return nil, nil
}

func (sr *SqlRunner) exec(ctx *RunCtx, opt *RunOptions) (*RunResult, error) {
	ch := make(chan *runtime.ExecResult, 1)

	query := fmt.Sprintf("%s", opt.Stdin)
	cmd := append(executeCmd, query)

	go sr.runtime.Exec(ctx.ContId, cmd, ch)

	var rr RunResult
	select {
	case res := <-ch:
		rr.ExitCode = res.ExitCode
		if res.ExitCode != 0 {
			rr.Status = Failed
			rr.Result = res.Stderr
		} else {
			rr.Status = Finished
			rr.Result = res.Stdout
		}
	case <-time.After(time.Duration(executionTimeLimit) * time.Millisecond):
		rr.Status = Interrupted
		go sr.runtime.Kill(ctx.ContId)
	}

	return &rr, nil
}

func (sr *SqlRunner) cleanUp(ctx *RunCtx) {
	sr.runtime.Kill(ctx.ContId)
}

func (sr *SqlRunner) Run(ctx *RunCtx, src string, options *RunOptions) *RunResult {

	var result RunResult
	result.Compiles = true
	result.ExitCode = -1
	result.Status = Failed

	options.Stdin = src
	res, err := sr.exec(ctx, options)
	if err != nil {
		result.Result = "Internal error"
	} else if res != nil {
		result.Result = res.Result
		result.ExitCode = res.ExitCode
		result.Status = Finished
	}

	return &result
}
