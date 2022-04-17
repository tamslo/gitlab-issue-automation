package standupNotes

import (
	dateUtils "gitlab-issue-automation/date_utils"
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	recurringIssues "gitlab-issue-automation/recurring_issues"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const StandupTitlePrefix = "Standup-Meetings/"

// tableColumns := []string{":rainbow: Project", ":back: What I did", ":soon: What I will do", ":warning:️ Problems", ":pencil: Notes" }

func WriteNotes(lastTime time.Time) {
	standupIssuePath := filepath.Join(gitlabUtils.GetRecurringIssuesPath(), "prepare-standup.md")
	_, err := os.Stat(standupIssuePath)
	standupIssueExists := err == nil
	if !standupIssueExists {
		return
	}
	standupIssue, err := recurringIssues.GetRecurringIssue(standupIssuePath, lastTime)
	if err != nil {
		log.Fatal(err)
	}
	// if standupIssue.NextTime.Before(time.Now()) {
	if standupIssue.NextTime.Before(time.Now()) || true {
		issueDue := gitlabUtils.GetIssueDueDate(standupIssue)
		issueDateString := issueDue.Format(dateUtils.ShortISODateLayout)
		issueDateString = strings.ReplaceAll(issueDateString, "-", "–")
		title := StandupTitlePrefix + issueDateString
		// if !gitlabUtils.WikiPageExists(title) {
		if !gitlabUtils.WikiPageExists(title) || true {
			// Collect relevant issues
			log.Println("Creating new wiki page", title)
			gitlabUtils.CreateWikiPage(title, "*This is YET another **test**.*")
		}
	}
}
