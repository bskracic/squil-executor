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
	Status string `json:"status"`
	Info   string `json:"info"`
	Table  Table  `json:"table"`
}

type Table struct {
	Columns []string            `json:"columns"`
	Rows    []map[string]string `json:"rows"`
	Info    string              `json:"info"`
}

const sessionName = "container-session"

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

var store = sessions.NewCookieStore([]byte("secret-key-mega"))

func main() {

	runtime := runtime.NewDockerRuntime()
	sqlRunner := runner.NewSqlRunner(runtime)

	router := mux.NewRouter()

	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {

		// vars := mux.Vars(r)
		// sessionID := vars["session_id"]

		// if sessionID == "" {
		// 	json.NewEncoder(w).Encode("no session :/")
		// } else {
		// 	json.NewEncoder(w).Encode("session!")
		// }

		id := GetContainerId(sqlRunner, w, r)
		json.NewEncoder(w).Encode(id)

	}).Methods("GET")

	router.HandleFunc("/exec", func(w http.ResponseWriter, r *http.Request) {

		var req request
		json.NewDecoder(r.Body).Decode(&req)

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
		session.Values["session_id"] = sr.CreateContainer()
		// Save session
		err = session.Save(r, w)
		if err != nil {
			return " "
		}
		return "new string needed"
	}
}
