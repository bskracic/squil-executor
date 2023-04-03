package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/bskracic/squil-executor/runner"
	"github.com/bskracic/squil-executor/runtime"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type request struct {
	Query string `json:"query"`
}

type response struct {
	Status  string        `json:"status"`
	Results []QueryResult `json:"queryResults"`
}

type QueryResult struct {
	Table Table  `json:"table"`
	Info  string `json:"info"`
}

type Table struct {
	Columns []string            `json:"columns"`
	Rows    []map[string]string `json:"rows"`
}

const sessionName = "container-session"

var store = sessions.NewCookieStore([]byte("secret-key-mega"))

func main() {

	runtime := runtime.NewDockerRuntime()
	sqlRunner := runner.NewSqlRunner(runtime)

	router := mux.NewRouter()

	router.HandleFunc("/exec", func(w http.ResponseWriter, r *http.Request) {

		var req request
		json.NewDecoder(r.Body).Decode(&req)

		id := GetContainerId(sqlRunner, w, r)

		ctx := &runner.RunCtx{
			ContId: id,
		}
		res := sqlRunner.Run(ctx, req.Query, &runner.RunOptions{})

		w.Header().Add("Content-Type", "application/json")

		json.NewEncoder(w).Encode(&response{
			Status:  string(res.Status),
			Results: prepareResults(strings.Split(res.Result, "\n")),
		})
	}).Methods("POST")

	log.Println("Listening on 1337")
	log.Fatal(http.ListenAndServe(":1337", router))
}

func GetContainerId(sr *runner.SqlRunner, w http.ResponseWriter, r *http.Request) string {
	session, err := store.Get(r, sessionName) // Replace with your own session name
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return ""
	}

	sessionId := session.Values["session_id"]

	if val, ok := sessionId.(string); ok {
		return val
	} else {
		id := sr.CreateContainer()
		session.Values["session_id"] = id
		// Save session
		err = session.Save(r, w)
		if err != nil {
			return " "
		}
		return id
	}
}

func prepareResults(lines []string) []QueryResult {
	var results []QueryResult
	for idx := 0; idx < len(lines); {
		var result QueryResult
		if strings.Contains(lines[idx], "-") { // we reached decorators, meaning that column names are before
			// remove previous result info, because it was definition of columns
			results = results[:len(results)-1]
			result.Table = extractTable(lines, &idx)
		}
		result.Info = lines[idx]
		results = append(results, result)
		idx++
	}

	results = results[:len(results)-1]

	return results
}

func extractTable(lines []string, idx *int) Table {
	columns := strings.Split(lines[*idx-1], ",") // columns are line before
	rows := []map[string]string{}
	var t Table
	for {
		*idx++
		line := lines[*idx]
		if len(line) == 0 {
			*idx++
			break
		}
		row := make(map[string]string)
		data := strings.Split(lines[*idx], ",")
		for index, point := range data {
			row[columns[index]] = point
		}
		rows = append(rows, row)
	}
	t.Columns = columns
	t.Rows = rows

	return t
}
