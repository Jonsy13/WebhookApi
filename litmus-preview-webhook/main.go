package main

import (
	client "WebHookSync/k8s-client"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

//namespace where, pods need to be rebooted.
const namespace = "litmus"

// struct pushData
type pushDataType struct {
	Images   []string `json:"images"`
	PushedAt int      `json:"pushed_at"`
	Pusher   string   `json:"pusher"`
	Tag      string   `json:"tag"`
}

// struct repositoryType
type repositoryType struct {
	CommentCount    int    `json:"comment_count"`
	DateCreated     int    `json:"date_created"`
	Description     string `json:"description"`
	Dockerfile      string `json:"dockerfile"`
	FullDescription string `json:"full_description"`
	IsOfficial      bool   `json:"is_official"`
	IsPrivate       bool   `json:"is_private"`
	IsTrusted       bool   `json:"is_trusted"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Owner           string `json:"owner"`
	RepoName        string `json:"repo_name"`
	RepoURL         string `json:"repo_url"`
	StarCount       int    `json:"start_count"`
	Status          string `json:"status"`
}

// Payload is a struct for payload coming from dockerhub
type Payload struct {
	ID          int64          `json:"id"`
	CallbackURL string         `json:"callback_url"`
	PushData    pushDataType   `json:"push_data"`
	Repository  repositoryType `json:"repository"`
}

var clientset *kubernetes.Clientset

//Homepage Handler
func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "homepage Endpoint Reached!!")
}

//Function for parsing the payload and deleting the
func webhookPayloadParser(w http.ResponseWriter, r *http.Request) {
	slackURL := os.Getenv("SLACK_WEBHOOK_LINK")
	reqBody, _ := ioutil.ReadAll(r.Body)
	var payload Payload
	json.Unmarshal(reqBody, &payload)
	postBody := "{\"text\":'" + payload.Repository.Name + " image was updated successfully.'}"
	klog.Infof("[Confirmation]: WebHook Hit Recieved: -----------------------------")
	klog.Infof("[Info]: Full RepoName : %v", payload.Repository.RepoName)
	klog.Infof("[Info]: Repository Name: %v", payload.Repository.Name)
	//Here, payload.Repository.Name is the name of image which is updated
	err := client.DeletePod(payload.Repository.Name, namespace, clientset)
	if err != nil {
		postBody = "{\"text\":' " + payload.Repository.Name + " image updation was unsuccessful, Error: " + err.Error() + "'}"
		klog.Errorf(err.Error())
	}

	_, err = http.Post(slackURL, "application/json", bytes.NewBuffer([]byte(postBody)))

	if err != nil {
		klog.Errorf("[Slack Confirmation]: Slack Notification was unsuccessful, err: %v", err)
	} else {
		klog.Infof("[Slack Confirmation]: Slack Notification sent successfully")
	}
}

//Function for handling requests
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", home).Methods("GET")
	myRouter.HandleFunc("/image-webhook", webhookPayloadParser).Methods("POST")
	log.Fatal(http.ListenAndServe(":8081", myRouter))
}

//Main Function
func main() {
	clientset, _ = client.GetGenericK8sClient()
	handleRequests()
}
