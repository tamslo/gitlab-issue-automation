package main

import (
	boardLabels "gitlab-issue-automation/board_labels"
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	recurringIssues "gitlab-issue-automation/recurring_issues"
	standupNotes "gitlab-issue-automation/standup_notes"
	"log"
	"time"
)

// TODO: Test biweekly occurance, adding labels, handling exceptions

func main() {
	lastRunTime := gitlabUtils.GetLastRunTime()
	log.Println("Last run:", lastRunTime.Format(time.RFC3339))
	log.Println("Creating recurring issues")
	recurringIssues.ProcessIssueFiles(lastRunTime)
	log.Println("Adapting board labels")
	boardLabels.AdaptLabels()
	log.Println("Creating standup notes")
	standupNotes.WriteNotes(lastRunTime)
	log.Println("Run complete")
}
