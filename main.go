package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/yztangent/planner/githubapi"
)

func main() {
	// Configuration
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable not set")
	}

	// These would eventually come from user config
	owner := "yztangent"
	projectNumberStr := "2"
	projectNumber, err := strconv.Atoi(projectNumberStr)
	if err != nil {
		log.Fatalf("Invalid project number: %v", err)
	}

	// Set a "since" timestamp. For this example, we'll use one year ago.
	since := time.Now().AddDate(-1, 0, 0)

	// Create the client
	ctx := context.Background()
	httpClient := &http.Client{}
	client := githubapi.NewClient(httpClient, githubToken)

	// Fetch the issues
	issues, err := client.GetIssues(ctx, owner, projectNumber, since)
	if err != nil {
		log.Fatalf("Failed to get issues: %v", err)
	}

	// Print the results
	fmt.Printf("Found %d issues in project %s/%d updated since %s:\n", len(issues), owner, projectNumber, since.Format("2006-01-02"))
	for _, issue := range issues {
		fmt.Printf("#%d: %s (Updated: %s)\n", issue.Number, issue.Title, issue.UpdatedAt.Format("2006-01-02"))
	}
}

