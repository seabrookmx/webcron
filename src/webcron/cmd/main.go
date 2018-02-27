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

	jobRoute := root.PathPrefix("/jobs").Subrouter()
	jobRoute.Path("/{id}").Methods("GET").HandlerFunc(getJob)
	jobRoute.Methods("POST").HandlerFunc(createJob)

	http.Handle("/", root)

	port := viper.GetInt("port")
	fmt.Printf("Listening on %d\n", port)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
