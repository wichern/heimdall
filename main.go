package main

// @todo: Provide disk usage stats https://gist.github.com/ttys3/21e2a1215cf1905ab19ddcec03927c75
// @todo: Provide Network status stats.
// @todo: Provide CPU/Memory usage stats.
// @todo: Provide local stylesheets and assets when available.
// @todo: User authentication with jwt.

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/wichern/heimdall/filebuffer"
	"github.com/wichern/heimdall/scriptrunner"
)

var files = filebuffer.Get()
var scripts = scriptrunner.Get("./data/examples")

func main() {
	port := *flag.String("port", "8080", "Port")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/status/{id}/{cmdIndex}", statusHandler)
	r.HandleFunc("/assets/{file}", assetHandler)
	r.HandleFunc("/api/scripts", getScriptsHandler).Methods("GET")
	r.HandleFunc("/api/scripts/start/{id}", startScriptHandler).Methods("POST")
	r.HandleFunc("/api/scripts/stop/{id}", stopScriptHandler).Methods("POST")
	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var indexTemplate, err = files.Get("./data/templates/index.tmpl")
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}

	type HTMLIndex struct {
		Title       string
		Stylesheets template.HTML
		Scripts     template.HTML
	}

	data := &HTMLIndex{
		Title: hostname,
		Stylesheets: `
		<link href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T" crossorigin="anonymous">
		<link href="https://stackpath.bootstrapcdn.com/font-awesome/4.7.0/css/font-awesome.min.css" rel="stylesheet" integrity="sha384-wvfXpqpZZVQGK6TAh5PVlGOfQNHSoD2xbE+QkPxCAFlNEevoEH3Sl0sibVcOQVnN" crossorigin="anonymous">`,
		Scripts: `
		<script src="http://code.jquery.com/jquery-3.4.1.min.js" integrity="sha256-CSXorXvZcTkaix6Yvo6HppcZGetbYMGWSFlBw8HfCJo="crossorigin="anonymous"></script>
		<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/js/bootstrap.min.js" integrity="sha384-JjSmVgyd0p3pXB1rRibZUAYoIIy6OrQ6VrjIEaFf/nJGzIxFDsf4x0xIM+B07jRM" crossorigin="anonymous"></script>
		<script src="assets/heimdall.js"></script>`,
	}

	if err := indexTemplate.GetTemplate().Execute(w, data); err != nil {
		http.NotFound(w, r)
		return
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	script, err := getScriptByID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	indexTemplate, err := files.Get("./data/templates/status.tmpl")
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}

	type HTMLStatus struct {
		Stdout string
	}

	vars := mux.Vars(r)
	index, err := strconv.Atoi(vars["cmdIndex"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stdout, err := script.Stdout(index)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := &HTMLStatus{
		Stdout: stdout,
	}

	if err := indexTemplate.GetTemplate().Execute(w, data); err != nil {
		http.NotFound(w, r)
		return
	}
}

func assetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	body, err := files.Get("./data/assets/" + vars["file"])
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}

	switch filepath.Ext(r.URL.Path) {
	case ".js":
		w.Header().Set("Content-Type", "text/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	default:
		// no content-type
	}

	w.Write(body.GetBody())
}

type scriptItem struct {
	ID      int
	Name    string
	Running bool
	Last    int // Index of last run
}

type ackMessage struct {
	status bool
}

func getScriptsHandler(w http.ResponseWriter, r *http.Request) {
	items := make([]scriptItem, len(scripts.Scripts))

	for index, element := range scripts.Scripts {
		if element.Running {
			fmt.Println("Process is running!")
		}
		items[index] = scriptItem{element.ID, element.Name, element.Running, element.LastRunIndex()}
	}

	json, err := json.Marshal(items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	}
}

func getScriptByID(r *http.Request) (*scriptrunner.Script, error) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return nil, err
	}

	if id < 0 || id >= len(scripts.Scripts) {
		return nil, errors.New("Script not found")
	}

	return &scripts.Scripts[id], nil
}

func startScriptHandler(w http.ResponseWriter, r *http.Request) {
	script, err := getScriptByID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = script.Start()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Script started")

	json, err := json.Marshal(ackMessage{true})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	}
}

func stopScriptHandler(w http.ResponseWriter, r *http.Request) {
	script, err := getScriptByID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = script.Stop()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Script aborted")

	json, err := json.Marshal(ackMessage{true})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("NotFoundHandler: " + r.URL.Path)
	http.NotFound(w, r)
}
