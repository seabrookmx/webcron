package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"webcron/jobs"

	"github.com/gorilla/mux"
)

func getJobs(w http.ResponseWriter, r *http.Request) {
	// TODO: interface with the job manager
	data, err := json.Marshal([]string{"things", "and", "stuff"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	io.WriteString(w, string(data[:]))
}

func createJob(w http.ResponseWriter, r *http.Request) {
	// TODO: interface with the job manager
	w.WriteHeader(http.StatusCreated)
}

func main() {
	jobs.NewJobManager()

	root := mux.NewRouter()

	jobRoute := root.PathPrefix("/jobs").Subrouter()
	jobRoute.Methods("GET").HandlerFunc(getJobs)
	jobRoute.Methods("POST").HandlerFunc(createJob)

	http.Handle("/", root)

	fmt.Println("Listening on 8080")
	http.ListenAndServe(":8080", nil)
}
