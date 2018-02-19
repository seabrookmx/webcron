package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"webcron/jobs"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var manager *jobs.JobManager

func getJob(w http.ResponseWriter, r *http.Request) {
	rawId := mux.Vars(r)["id"]

	id, err := uuid.Parse(rawId)
	var job jobs.Job
	if err == nil {
		job = manager.GetJob(id)
	}
	var encodedJob []byte
	if err == nil {
		encodedJob, err = json.Marshal(job)
	}
	if err != nil {
		http.Error(w, "Error occurred retrieving job", http.StatusInternalServerError)
		return
	}
	if job.Id == uuid.Nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	io.WriteString(w, string(encodedJob[:]))
}

func createJob(w http.ResponseWriter, r *http.Request) {
	var job jobs.Job

	bytes, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(bytes, &job)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON body: %s", err), http.StatusBadRequest)
		return
	}

	job, err = manager.CreateJob(job)
	var encodedJob []byte
	if err == nil {
		encodedJob, err = json.Marshal(job)
	}
	if err != nil {
		http.Error(w, "Error creating job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, string(encodedJob[:]))
}

func main() {
	manager = jobs.NewJobManager()

	root := mux.NewRouter()

	jobRoute := root.PathPrefix("/jobs").Subrouter()
	jobRoute.Path("/{id}").Methods("GET").HandlerFunc(getJob)
	jobRoute.Methods("POST").HandlerFunc(createJob)

	http.Handle("/", root)

	fmt.Println("Listening on 8080")
	http.ListenAndServe(":8080", nil)
}
