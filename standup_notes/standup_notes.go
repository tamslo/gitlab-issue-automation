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
	verbose := false
	standupIssue, err := recurringIssues.GetRecurringIssue(standupIssuePath, lastTime, verbose)
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
			orderBy := "updated_at"
			sortOrder := "asc"
			openIssues := gitlabUtils.GetSortedProjectIssues(orderBy, sortOrder, "opened")
			closedIssues := gitlabUtils.GetSortedProjectIssues(orderBy, sortOrder, "closed")
			for _, issue := range openIssues {
				log.Println(issue.UpdatedAt)
			}
			for _, issue := range closedIssues {
				log.Println(issue.UpdatedAt)
			}
			content := "*This is YET another **test**.*"
			log.Println("Would create new wiki page", title, "with content", content)
			// gitlabUtils.CreateWikiPage(title, content)
		}
	}
}
