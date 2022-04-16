package thisWeekLabel

import (
	dateUtils "gitlab-issue-automation/date_utils"
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	"log"
	"time"

	"github.com/xanzy/go-gitlab"
)

const ThisWeekLabel = "🗓 This week"

var OtherLabels = [...]string{"🏢 In office", "🏃‍♀️ In progress", "⏳ Waiting"}

func AdaptLabels() {
	git := gitlabUtils.GetGitClient()
	project := gitlabUtils.GetGitProject()
	issueState := "opened"
	orderBy := "due_date"
	sortOrder := "asc"
	perPage := 20
	page := 1
	lastPageReached := false
	var issues []*gitlab.Issue
	for {
		if lastPageReached {
			break
		}
		listOptions := &gitlab.ListOptions{
			PerPage: perPage,
			Page:    page,
		}
		options := &gitlab.ListProjectIssuesOptions{
			State:       &issueState,
			OrderBy:     &orderBy,
			Sort:        &sortOrder,
			ListOptions: *listOptions,
		}
		pageIssues, _, err := git.Issues.ListProjectIssues(project.ID, options)
		if err != nil {
			log.Fatal(err)
		}
		issues = append(issues, pageIssues...)
		if len(pageIssues) < perPage {
			lastPageReached = true
		} else {
			page++
		}
	}
	for _, issue := range issues {
		if issue.DueDate == nil {
			continue
		}
		issueDueTime, err := time.Parse(dateUtils.ShortISODateLayout, issue.DueDate.String())
		if err != nil {
			log.Fatal(err)
		}
		issueDueWeekStart := dateUtils.GetStartOfWeek(issueDueTime)
		currentWeekStart := dateUtils.GetStartOfWeek(time.Now())
		issueDueThisWeek := issueDueWeekStart.Before(currentWeekStart) ||
			dateUtils.AreDatesEqual(issueDueWeekStart, currentWeekStart)
		if !issueDueThisWeek {
			break
		}
		issueHasNextWeekLabel := false
		issueHasOtherLabel := false
		for _, label := range issue.Labels {
			if issueHasNextWeekLabel || issueHasOtherLabel {
				break
			}
			if label == ThisWeekLabel {
				issueHasNextWeekLabel = true
			}
			for _, otherLabel := range OtherLabels {
				if label == otherLabel {
					issueHasOtherLabel = true
					break
				}
			}
		}
		issueNeedsThisWeekLabel := !issueHasNextWeekLabel && !issueHasOtherLabel
		if issueNeedsThisWeekLabel {
			issueName := "'" + issue.Title + "'"
			log.Println("Moving issue", issueName, "to this week")
			updatedLabels := append(issue.Labels, ThisWeekLabel)
			options := &gitlab.UpdateIssueOptions{
				Labels: &updatedLabels,
			}
			_, _, err := git.Issues.UpdateIssue(project.ID, issue.IID, options)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
