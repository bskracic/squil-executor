package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/bskracic/squil-executor/runner"
	"github.com/bskracic/squil-executor/runtime"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
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

func main() {

	runtime := runtime.NewDockerRuntime()
	sqlRunner := runner.NewSqlRunner(runtime)

	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type"}
	router.Use(cors.New(config))

	// Initialize cookie store
	store := cookie.NewStore([]byte("dugacki-secret"))

	// Set session options
	store.Options(sessions.Options{
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteNoneMode,
	})

	// Use the cookie store for sessions
	router.Use(sessions.Sessions(sessionName, store))

	router.POST("/api/query", func(c *gin.Context) {
		var req request
		json.NewDecoder(c.Request.Body).Decode(&req)

		// id := GetContainerId(sqlRunner, c)
		ctx := &runner.RunCtx{
			ContId: "ea16a132713364d8b0796541d07ab09e7374aed92f8e56273105096dcb912c4b",
		}
		res := sqlRunner.Run(ctx, req.Query, &runner.RunOptions{})

		c.Writer.Header().Set("Access-Control-Expose-Headers", "Set-Cookie")

		log.Println(res.Result)

		c.JSON(200, &response{
			Status:  string(res.Status),
			Results: prepareResults(strings.Split(res.Result, "\n")),
		})
	})

	log.Println("Server starting on port 1337...")
	router.Run(":1337")
}

func GetContainerId(sr *runner.SqlRunner, c *gin.Context) string {

	session := sessions.Default(c)
	id := session.Get("session_id")
	if id == nil {
		id := sr.CreateContainer()
		log.Printf("created container: %v\n", id)
		session.Set("session_id", id)
		session.Save()
		return id
	} else if value, ok := id.(string); ok {
		return value
	} else {
		c.AbortWithStatus(http.StatusInternalServerError)
		return ""
	}
}

func prepareResults(lines []string) []QueryResult {
	var results []QueryResult

	for idx := 0; idx < len(lines); {
		var result QueryResult
		if strings.HasPrefix(lines[idx], "-") { // we reached decorators, meaning that column names are before
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
