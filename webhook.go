package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	api "code.gitea.io/gitea/modules/structs"
	"gopkg.in/yaml.v3"
)

type TaskToRun struct {
	Namespace               string
	ServiceAccount          string
	TaskRefName             string
	GitSourceResourceName   string
	DockerImageResourceName string
}

type HookAction struct {
	Secret   string
	Repo     string
	TaskInfo TaskToRun
}

type Config struct {
	Hooks []HookAction
}

func readConfig(filename string) (Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}

	var conf Config
	yaml.Unmarshal(content, &conf)
	log.Printf("%#v", conf)

	return conf, nil
}

func action(hook HookAction) {
	log.Printf("exec command for repo:%s", hook.Repo)
	ti := hook.TaskInfo
	err := CreateTaskRun(ti.Namespace, ti.ServiceAccount, ti.TaskRefName, ti.GitSourceResourceName, ti.DockerImageResourceName)
	if err != nil {
		log.Printf("Failed to create taskrun:%v", err)
	}
}

func webhook(w http.ResponseWriter, r *http.Request) {
	event := r.Header.Get("X-Gitea-Event")

	if event != "push" {
		log.Printf("unknown event:%s", event)
		http.Error(w, "Unsupported event", 400)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("body read failed: %v", err)
		http.Error(w, "Fail to read body", 500)
		return
	}

	var payload api.PushPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Printf("Unmarshal body failed:%v", err)
		http.Error(w, "failed to parse payload", 500)
		return
	}

	for _, hook := range config.Hooks {
		log.Printf("repo=%s,payload repo=%s, secret=%s, payload secret=%s", hook.Repo, payload.Repo.FullName, hook.Secret, payload.Secret)
		if hook.Repo == payload.Repo.FullName && hook.Secret == payload.Secret {
			go action(hook)
			fmt.Fprintf(w, "Push event dispatched")
			return
		}
	}
	http.Error(w, "not matched.", 404)
}

var config Config

func main() {
	var err error
	config, err = readConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read config yaml:%v", err)
	}

	http.HandleFunc("/webhook", webhook)
	port := os.Getenv("PORT")
	if port == "" {
		port = "9691"
	}
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
