package standupNotes

import (
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	recurringIssues "gitlab-issue-automation/recurring_issues"
	"log"
	"os"
	"path/filepath"
	"time"
)

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
		log.Println("TODO: Create standup notes")
		//issueDue := gitlabUtils.GetIssueDueDate(standupIssue)
		//title := issueDue.Format(dateUtils.PrettyDateLayout)
		// Create Wiki page
		// Collect relevant issues
		// TODO: What happens if wiki page with same name exists
		gitlabUtils.CreateWikiPage("Test", "*This is another **test**.*")
	}
}
