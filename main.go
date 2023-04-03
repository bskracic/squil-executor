package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/bskracic/squil-executor/runner"
	"github.com/bskracic/squil-executor/runtime"
)

type request struct {
	Query string `json:"query"`
}

type response struct {
	Status string `json:"status"`
	Info   string `json:"info"`
	Table  Table  `json:"table"`
}

type Table struct {
	Columns []string            `json:"columns"`
	Rows    []map[string]string `json:"rows"`
	Info    string              `json:"info"`
}

func prepareTables(resultText string) Table {
	text := resultText
	lines := strings.Split(text, "\n")

	var t Table
	columns := strings.Split(lines[0], ",")
	rows := []map[string]string{}

	// first line are the column names
	// skip the decorative line
	var i int
	for i = 2; i < len(lines); i++ {
		if len(lines[i]) == 0 {
			i++
			break
		}
		row := make(map[string]string)
		data := strings.Split(lines[i], ",")
		for index, point := range data {
			row[columns[index]] = point
		}
		rows = append(rows, row)
	}
	t.Info = strings.Join(lines[i:], "\n")
	t.Columns = columns
	t.Rows = rows

	return t
}

func main() {

	runtime := runtime.NewDockerRuntime()
	sqlRunner := runner.NewSqlRunner(runtime)

	http.HandleFunc("/exec", func(w http.ResponseWriter, r *http.Request) {

		var req request
		json.NewDecoder(r.Body).Decode(&req)
		// retrieve container id from session id / reddis
		// if there is none, create one
		ctx := &runner.RunCtx{
			ContId: "81d5a95712fc650ff94cc77b9197d3ee7c2f0e6e1b1fea64d9eeada25a2ba95b",
		}
		rr := sqlRunner.Run(ctx, req.Query, &runner.RunOptions{})

		var res response
		if rr.ExitCode == 0 && rr.Status == runner.Finished {
			res.Table = prepareTables(rr.Result)
		} else {
			res.Info = rr.Result
		}
		res.Status = string(rr.Status)

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	})

	log.Println("Listening on 1337")
	log.Fatal(http.ListenAndServe(":1337", nil))
}
