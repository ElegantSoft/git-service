package main

import (
	"fmt"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
	// Create a new GitHub webhook instance
	hook, _ := github.New(github.Options.Secret(os.Getenv("WEB_HOOK_SECRET")))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("got /hello request\n")
		io.WriteString(w, "Hello, HTTP!\n")
	})

	// Start the server and listen for incoming GitHub webhook events
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.PushEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				// Ignore requests that are not for push events
				return
			}
			log.Printf("Failed to parse webhook payload: %s", err)
			return
		}

		// Handle the parsed webhook payload
		switch payload.(type) {

		case github.PushPayload:
			pushEvent := payload.(github.PushPayload)
			fmt.Printf("%+v", pushEvent)
			// Do whatever you want from here...
			if pushEvent.Ref == fmt.Sprintf("refs/heads/%v", os.Getenv("BRANCH_NAME")) && pushEvent.Repository.FullName == os.Getenv("REPO_NAME") {
				// Run the specific .sh file
				cmd := exec.Command("bash", os.Getenv("SHELL_PATH"))
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					log.Printf("Failed to run script: %s", err)
				}
			}

		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)
			// Do whatever you want from here...
			fmt.Printf("%+v", pullRequest)
		}
	})

	log.Printf("Starting server on : %v...", os.Getenv("PORT"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", os.Getenv("port")), nil))
}
