package main

import (
	boardLabels "gitlab-issue-automation/board_labels"
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	recurringIssues "gitlab-issue-automation/recurring_issues"
	standupNotes "gitlab-issue-automation/standup_notes"
	"log"
	"time"
)

func main() {
	lastRunTime := gitlabUtils.GetLastRunTime()
	log.Println("Last run:", lastRunTime.Format(time.RFC3339))
	log.Println("Checking whether to create recurring issues")
	recurringIssues.ProcessIssueFiles(lastRunTime)
	log.Println("Checking whether to adapt board labels")
	boardLabels.AdaptLabels()
	log.Println("Checking whether to create standup notes")
	standupNotes.WriteNotes(lastRunTime)
	log.Println("Run complete")
}
