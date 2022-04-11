package main

import (
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	"log"
	"path/filepath"
	"time"
)

// TODO: Test biweekly occurance, adding labels, handling exceptions

func main() {
	lastRunTime, err := gitlabUtils.GetLastRunTime()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Last run:", lastRunTime.Format(time.RFC3339))

	log.Println("Creating recurring issues")
	err = filepath.Walk(recurringIssues.getRecurringIssuesPath(), recurringIssues.processIssueFile(lastRunTime))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Adapting labels")

	err = adaptLabels()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Run complete")
}
