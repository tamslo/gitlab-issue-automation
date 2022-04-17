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

	"github.com/xanzy/go-gitlab"
)

const StandupTitlePrefix = "Standup-Meetings/"
const enDash = "–"
const dash = "-"
const lookupStart = "2022-04-06"

func escapeDashes(text string) string {
	return strings.ReplaceAll(text, dash, enDash)
}

func unescapeDashes(text string) string {
	return strings.ReplaceAll(text, enDash, dash)
}

func getLastNoteDate(currentDate time.Time) time.Time {
	git := gitlabUtils.GetGitClient()
	project := gitlabUtils.GetGitProject()
	options := &gitlab.ListWikisOptions{}
	wikiPages, _, err := git.Wikis.ListWikis(project.ID, options)
	if err != nil {
		log.Fatal(err)
	}
	latestStandup, err := time.Parse(dateUtils.ShortISODateLayout, lookupStart)
	if err != nil {
		log.Fatal(err)
	}
	for _, wikiPage := range wikiPages {
		if !strings.HasPrefix(wikiPage.Title, StandupTitlePrefix) {
			continue
		}
		thisStandupDate, err := time.Parse(dateUtils.ShortISODateLayout, unescapeDashes(strings.Replace(wikiPage.Title, StandupTitlePrefix, "", 1)))
		if err != nil {
			log.Fatal(err)
		}
		if thisStandupDate.After(latestStandup) {
			latestStandup = thisStandupDate
		}
	}
	return latestStandup
}

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
		issueDueString := escapeDashes(issueDue.Format(dateUtils.ShortISODateLayout))
		title := StandupTitlePrefix + issueDueString
		// if !gitlabUtils.WikiPageExists(title) {
		if !gitlabUtils.WikiPageExists(title) || true {
			content := "| :rainbow: Project | :back: What I did | :soon: What I will do | :warning:️ Problems | :pencil: Notes |\n"
			content += "|-------------------|-------------------|-----------------------|--------------------|----------------|\n"
			lastNoteDate := getLastNoteDate(issueDue)
			orderBy := "updated_at"
			sortOrder := "desc"
			issues := gitlabUtils.GetSortedProjectIssues(orderBy, sortOrder, "")
			content += "\n"
			content += "## Issues\n"
			content += "\n"
			for _, issue := range issues {
				log.Println(issue.Title)
				log.Println(issue.UpdatedAt)
				if issue.UpdatedAt.After(lastNoteDate) {
					content += "* [" + issue.Title + "]"
					content += "(" + issue.WebURL + ")"
					content += " \\[" + strings.Join(append(issue.Labels, issue.State), ", ") + "\\]"
				}
			}
			log.Println("Would create new wiki page", title, "with content", content)
			// gitlabUtils.CreateWikiPage(title, content)
		} else {
			log.Println("Skipping creation of wiki page", title, "because it already exists")
		}
	}
}
