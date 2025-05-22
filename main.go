package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/joho/godotenv"
)

// Target represents a repository, branch and script to execute
type Target struct {
	RepoName  string
	Branch    string
	ShellPath string
}

// parseTargets parses the environment variables to get multiple targets
func parseTargets() []Target {
	var targets []Target

	// Get the number of targets
	count := 1
	for {
		repoName := os.Getenv(fmt.Sprintf("REPO_NAME_%d", count))
		if repoName == "" {
			break
		}

		branch := os.Getenv(fmt.Sprintf("BRANCH_NAME_%d", count))
		shellPath := os.Getenv(fmt.Sprintf("SHELL_PATH_%d", count))

		// Only add if all required fields are present
		if branch != "" && shellPath != "" {
			targets = append(targets, Target{
				RepoName:  repoName,
				Branch:    branch,
				ShellPath: shellPath,
			})
		}

		count++
	}

	// Handle legacy single target format if no numbered targets found
	if len(targets) == 0 {
		repoName := os.Getenv("REPO_NAME")
		branch := os.Getenv("BRANCH_NAME")
		shellPath := os.Getenv("SHELL_PATH")

		if repoName != "" && branch != "" && shellPath != "" {
			targets = append(targets, Target{
				RepoName:  repoName,
				Branch:    branch,
				ShellPath: shellPath,
			})
		}
	}

	return targets
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	// Parse targets from environment variables
	targets := parseTargets()

	if len(targets) == 0 {
		log.Println("Warning: No targets configured")
	} else {
		log.Printf("Loaded %d target(s)", len(targets))
	}

	// Create a new GitHub webhook instance
	hook, _ := github.New(github.Options.Secret(os.Getenv("WEB_HOOK_SECRET")))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

			// Check against all targets
			for _, target := range targets {
				refBranch := fmt.Sprintf("refs/heads/%v", target.Branch)

				if pushEvent.Ref == refBranch && pushEvent.Repository.FullName == target.RepoName {
					log.Printf("Matched target: repo=%s, branch=%s", target.RepoName, target.Branch)

					// Run the specific .sh file
					cmd := exec.Command("bash", target.ShellPath)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Run()
					if err != nil {
						log.Printf("Failed to run script %s: %s", target.ShellPath, err)
					} else {
						log.Printf("Successfully executed script: %s", target.ShellPath)
					}
				}
			}

		case github.PullRequestPayload:
			pullRequest := payload.(github.PullRequestPayload)
			// Do whatever you want from here...
			fmt.Printf("%+v", pullRequest)
		}
		io.WriteString(w, "Hello, HTTP!\n")
	})
	http.HandleFunc("/hello", getHello)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9092" // Default port if not specified
	}

	log.Printf("Starting server on port: %v...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	io.WriteString(w, "Hello, HTTP!\n")
}
