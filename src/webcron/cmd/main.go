package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"webcron/dto"
	"webcron/jobs"
	"webcron/postbacks"

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
	if job.ID == "" {
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
	var job dto.Job

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
		errString := err.Error()
		if strings.Contains(errString, "Failure adding") {
			http.Error(w, errString, http.StatusInternalServerError)
			return
		}

		http.Error(w, errString, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	io.WriteString(w, string(encodedJob[:]))
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("wc") // ex: WC_DRYRUN
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err == nil {
		var postbackTrigger postbacks.PostbackTrigger
		if viper.GetBool("dryRun") {
			postbackTrigger = postbacks.NewLogPostbackTrigger()
		} else {
			postbackTrigger = postbacks.NewHTTPPostbackTrigger()
		}

		manager, err = jobs.NewJobManager(postbackTrigger)
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
