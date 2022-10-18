package standupNotes

import (
	"fmt"
	boardLabels "gitlab-issue-automation/board_labels"
	constants "gitlab-issue-automation/constants"
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

const StandupTitlePrefix = "Meetings/Standup/"
const lookupStart = "2022-04-06"

func getLastNoteDate(currentDate time.Time) time.Time {
	wikiPages := gitlabUtils.GetWikiPagesMetadata()
	latestStandup, err := time.Parse(dateUtils.ShortISODateLayout, lookupStart)
	if err != nil {
		log.Fatal(err)
	}
	for _, wikiPage := range wikiPages {
		if !strings.HasPrefix(wikiPage.Slug, StandupTitlePrefix) {
			continue
		}
		if !dateUtils.IsDashedDate(wikiPage.Title) {
			continue
		}
		thisStandupDate, err := time.Parse(dateUtils.ShortISODateLayout, dateUtils.UnescapeDashes(wikiPage.Title))
		if err != nil {
			log.Fatal(err)
		}
		if thisStandupDate.After(latestStandup) {
			latestStandup = thisStandupDate
		}
	}
	return latestStandup
}

func getComparableLabels(issue *gitlab.Issue) string {
	labels := issue.Labels
	sort.Strings(labels)
	return strings.Join(labels, ", ")
}

func printIssue(issue *gitlab.Issue) string {
	issueString := "* [ ] "
	if issue.State == "closed" {
		issueString += "✅ "
	} else if issue.State == "opened" {
		issueString += "📝 "
	} else {
		issueString += "❓ "
	}
	issueString += "[#" + fmt.Sprint(issue.IID) + " " + issue.Title + "]"
	issueString += "(" + issue.WebURL + ")"
	issueString += " (" + getComparableLabels(issue) + ")\n"
	return issueString
}

func WriteNotes(lastTime time.Time) {
	standupIssuePath := filepath.Join(gitlabUtils.GetRecurringIssuesPath(), constants.StandupIssueTemplateName)
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
		issueDueString := dateUtils.GetEnDashDate(issueDue)
		title := StandupTitlePrefix + issueDueString
		if !gitlabUtils.WikiPageExists(title) {
			lastNoteDate := getLastNoteDate(issueDue)
			orderBy := "updated_at"
			sortOrder := "desc"
			issues := gitlabUtils.GetSortedProjectIssues(orderBy, sortOrder, "")
			relevantIssues := []*gitlab.Issue{}
			projects := []string{}
			for _, issue := range issues {
				if boardLabels.HasLabel(issue, constants.TestLabel) || boardLabels.HasLabel(issue, constants.RecurringLabel) {
					continue
				}
				if issue.UpdatedAt.After(lastNoteDate) {
					if issue.State != "closed" || (issue.State == "closed" && issue.ClosedAt.After(lastNoteDate)) {
						relevantIssues = append(relevantIssues, issue)
						projectLabels := []string{}
						for _, label := range issue.Labels {
							isNonProjectLabel := true
							for _, nonProjectLabel := range constants.NonProjectLabels {
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

			sort.Slice(relevantIssues, func(firstIndex, secondIndex int) bool {
				firstLabels := getComparableLabels(relevantIssues[firstIndex])
				secondLabels := getComparableLabels(relevantIssues[secondIndex])
				return firstLabels < secondLabels
			})
			for _, issue := range relevantIssues {
				content += printIssue(issue)
			}
			log.Println("- Creating new wiki page", title)
			gitlabUtils.CreateWikiPage(title, content)
		} else {
			log.Println("- Skipping creation of wiki page", title, "because it already exists")
		}
	}
}
