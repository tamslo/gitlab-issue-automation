package boardLabels

import (
	constants "gitlab-issue-automation/constants"
	dateUtils "gitlab-issue-automation/date_utils"
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	"log"
	"time"

	"github.com/xanzy/go-gitlab"
)

func hasAnyLabel(issue *gitlab.Issue, labels []string) bool {
	anyLabelPresent := false
	for _, label := range labels {
		labelPresent := HasLabel(issue, label)
		if labelPresent {
			anyLabelPresent = true
			break
		}
	}
	return anyLabelPresent
}

func HasLabel(issue *gitlab.Issue, wantedLabel string) bool {
	labelPresent := false
	for _, label := range issue.Labels {
		if label == wantedLabel {
			labelPresent = true
			break
		}
	}
	return labelPresent
}

func removeLabel(issue *gitlab.Issue, unwantedLabel string) *gitlab.Issue {
	action := "Removing"
	preposition := "from"
	updatedLabels := gitlab.Labels{}
	for _, label := range issue.Labels {
		if !(label == unwantedLabel) {
			updatedLabels = append(updatedLabels, label)
		}
	}
	return adaptLabel(issue, unwantedLabel, updatedLabels, action, preposition)
}

func addLabel(issue *gitlab.Issue, label string) *gitlab.Issue {
	action := "Moving"
	preposition := "to"
	updatedLabels := append(issue.Labels, label)
	return adaptLabel(issue, label, updatedLabels, action, preposition)
}

func adaptLabel(issue *gitlab.Issue, label string, updatedLabels gitlab.Labels, action string, preposition string) *gitlab.Issue {
	issueName := "'" + issue.Title + "'"
	log.Println("-", action, "issue", issueName, preposition, label)
	options := &gitlab.UpdateIssueOptions{
		Labels: updatedLabels,
	}
	return gitlabUtils.UpdateIssue(issue.IID, options)
}

func AdaptLabels() {
	orderBy := "due_date"
	sortOrder := "asc"
	issueState := "opened"
	issues := gitlabUtils.GetSortedProjectIssues(orderBy, sortOrder, issueState)
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
		issuePastDue := issueDueTime.Before(time.Now())
		issueDueToday := dateUtils.AreDatesEqual(issueDueTime, time.Now())
		issueDueThisWeek := dateUtils.AreDatesEqual(issueDueWeekStart, currentWeekStart)
		if !(issuePastDue || issueDueToday || issueDueThisWeek) {
			break
		}
		issueHasProgressLabel := hasAnyLabel(issue, constants.ProgressLabels)
		if !issueHasProgressLabel {
			issueHasTodayLabel := HasLabel(issue, constants.TodayLabel)
			issueHasNextWeekLabel := HasLabel(issue, constants.ThisWeekLabel)
			if (issuePastDue || issueDueToday) && !issueHasTodayLabel {
				issue = addLabel(issue, constants.TodayLabel)
				if issueHasNextWeekLabel {
					removeLabel(issue, constants.ThisWeekLabel)
				}
			} else if !issueHasTodayLabel && issueDueThisWeek && !issueHasNextWeekLabel {
				addLabel(issue, constants.ThisWeekLabel)
			}
		}
	}
}

func CleanLabels(lastRunTime time.Time) {
	orderBy := "updated_at"
	sortOrder := "desc"
	issueState := "closed"
	issues := gitlabUtils.GetSortedProjectIssues(orderBy, sortOrder, issueState)
	for _, issue := range issues {
		log.Println("-- [DEBUG] Updated at:", issue.UpdatedAt)
		if issue.UpdatedAt.Before(lastRunTime) {
			log.Println("-- [DEBUG] Would break now")
			// break
		}
		for _, statusLabel := range constants.StatusLabels {
			if HasLabel(issue, statusLabel) {
				removeLabel(issue, statusLabel)
			}
		}
	}
}
