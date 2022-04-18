package standupNotes

import (
	"fmt"
	boardLabels "gitlab-issue-automation/board_labels"
	dateUtils "gitlab-issue-automation/date_utils"
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	recurringIssues "gitlab-issue-automation/recurring_issues"
	"log"
	"os"
	"path/filepath"
	"sort"
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

func printIssue(issue *gitlab.Issue) string {
	issueString := "* [#" + fmt.Sprint(issue.IID) + " " + issue.Title + "]"
	issueString += "(" + issue.WebURL + ")"
	issueString += " \\[" + strings.Join(append(issue.Labels, issue.State), ", ") + "\\]\n"
	return issueString
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
	if standupIssue.NextTime.Before(time.Now()) {
		issueDue := gitlabUtils.GetIssueDueDate(standupIssue)
		issueDueString := escapeDashes(issueDue.Format(dateUtils.ShortISODateLayout))
		title := StandupTitlePrefix + issueDueString
		if !gitlabUtils.WikiPageExists(title) {
			lastNoteDate := getLastNoteDate(issueDue)
			orderBy := "updated_at"
			sortOrder := "desc"
			issues := gitlabUtils.GetSortedProjectIssues(orderBy, sortOrder, "")
			relevantIssues := []*gitlab.Issue{}
			projects := []string{}
			for _, issue := range issues {
				if boardLabels.HasLabel(issue, boardLabels.TestLabel) || boardLabels.HasLabel(issue, boardLabels.RecurringLabel) {
					continue
				}
				if issue.UpdatedAt.After(lastNoteDate) {
					relevantIssues = append(relevantIssues, issue)
					projectLabels := []string{}
					for _, label := range issue.Labels {
						isNonProjectLabel := true
						for _, nonProjectLabel := range boardLabels.NonProjectLabels {
							if label == nonProjectLabel {
								isNonProjectLabel = false
								break
							}
						}
						if isNonProjectLabel {
							projectLabels = append(projectLabels, label)
						}
					}
					for _, label := range projectLabels {
						labelInProjects := false
						for _, project := range projects {
							if label == project {
								labelInProjects = true
								break
							}
						}
						if !labelInProjects {
							projects = append(projects, label)
						}
					}
				}
			}
			content := "| :rainbow: Project | :back: What I did | :soon: What I will do | :warning:️ Problems | :pencil: Notes |\n"
			content += "|-------------------|-------------------|-----------------------|--------------------|----------------|\n"
			sort.Strings(projects)
			for _, project := range projects {
				content += "| " + project + " |  |  |  |  |\n"
			}
			content += "\n"
			content += "## Issues\n"
			content += "\n"
			coveredIssueIds := []int{}
			for _, project := range projects {
				content += "### " + project + "\n"
				content += "\n"
				for _, issue := range relevantIssues {
					if boardLabels.HasLabel(issue, project) {
						coveredIssueIds = append(coveredIssueIds, issue.ID)
						content += printIssue(issue)
					}
				}
			}
			content += "\n"
			content += "### Non-project issues\n"
			content += "\n"
			allIssuesCovered := true
			for _, issue := range relevantIssues {
				issueCovered := false
				for _, issueId := range coveredIssueIds {
					if issueId == issue.ID {
						issueCovered = true
						break
					}
				}
				if !issueCovered {
					allIssuesCovered = false
					content += printIssue(issue)
				}
			}
			if allIssuesCovered {
				content += "_No non-project issues present_"
			}
			log.Println("Creating new wiki page", title)
			gitlabUtils.CreateWikiPage(title, content)
		} else {
			log.Println("Skipping creation of wiki page", title, "because it already exists")
		}
	}
}
