package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"webcron/jobs"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var manager *jobs.JobManager

func getJob(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	job := manager.GetJob(id)
	encodedJob, err := json.Marshal(job)
	if err != nil {
		http.Error(w, "Error occurred retrieving job", http.StatusInternalServerError)
		return
	}
	if job.Id == "" {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	io.WriteString(w, string(encodedJob[:]))
}

func getJobs(w http.ResponseWriter, r *http.Request) {
	skipStr := r.FormValue("skip")
	if skipStr == "" {
		skipStr = "0"
	}
	skip, _ := strconv.Atoi(skipStr)

	limitStr := r.FormValue("limit")
	if limitStr == "" {
		limitStr = "100"
	}
	limit, _ := strconv.Atoi(limitStr)

	jobs := manager.GetJobs(limit, skip)
	encodedJobs, err := json.Marshal(jobs)
	if err != nil {
		http.Error(w, "Error occurred retrieving job", http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(encodedJobs[:]))
}

func removeJob(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	err := manager.RemoveJob(id)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
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
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	// TODO: parse environment variables here
	err := viper.ReadInConfig()
	if err == nil {
		// TODO: handle more configuration knobs here
		manager, err = jobs.NewJobManager(viper.GetBool("dryRun"))
	}
	if err != nil {
		panic(errors.Wrap(err, "Error initializing webcron"))
	}

	root := mux.NewRouter()
	rootRoute := root.PathPrefix("/jobs").Subrouter()
	jobRoute := rootRoute.Path("").Subrouter()
	jobRoute.Methods("POST").HandlerFunc(createJob)
	jobRoute.Methods("GET").HandlerFunc(getJobs)

	jobIDRoute := rootRoute.Path("/{id}").Subrouter()
	jobIDRoute.Methods("GET").HandlerFunc(getJob)
	jobIDRoute.Methods("DELETE").HandlerFunc(removeJob)

	http.Handle("/", root)

	port := viper.GetInt("port")
	fmt.Printf("Listening on %d\n", port)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
