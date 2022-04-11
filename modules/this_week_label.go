package thisWeekLabel

import (
	"log"
	"time"

	"github.com/xanzy/go-gitlab"
)

func adaptLabels() error {
	git, err := getGitClient()
	if err != nil {
		return err
	}
	project, err := getGitProject(git)
	if err != nil {
		return err
	}
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
			return err
		}
		issues = append(issues, pageIssues...)
		if len(pageIssues) < perPage {
			lastPageReached = true
		} else {
			page++
		}
	}
	thisWeekLabel := "🗓 This week"
	otherLabels := []string{"🏢 In office", "🏃‍♀️ In progress", "⏳ Waiting"}
	for _, issue := range issues {
		if issue.DueDate == nil {
			continue
		}
		issueDueTime, err := time.Parse(shortISODateLayout, issue.DueDate.String())
		if err != nil {
			return err
		}
		issueDueWeekStart := getStartOfWeek(issueDueTime)
		currentWeekStart := getStartOfWeek(time.Now())
		issueDueThisWeek := issueDueWeekStart.Before(currentWeekStart) ||
			areDatesEqual(issueDueWeekStart, currentWeekStart)
		if !issueDueThisWeek {
			break
		}
		issueHasNextWeekLabel := false
		issueHasOtherLabel := false
		for _, label := range issue.Labels {
			if issueHasNextWeekLabel || issueHasOtherLabel {
				break
			}
			if label == thisWeekLabel {
				issueHasNextWeekLabel = true
			}
			for _, otherLabel := range otherLabels {
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
			updatedLabels := append(issue.Labels, thisWeekLabel)
			options := &gitlab.UpdateIssueOptions{
				Labels: &updatedLabels,
			}
			_, _, err := git.Issues.UpdateIssue(project.ID, issue.IID, options)
			if err != nil {
				return err
			}
		}
	}
	return err
}
