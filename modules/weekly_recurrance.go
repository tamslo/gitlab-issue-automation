package weeklyRecurrance

import (
	"math"
	"time"

	"github.com/xanzy/go-gitlab"
)

func checkWeeklyRecurrance(nextTime time.Time, data *types.metadata) time.Time {
	if data.WeeklyRecurrence > 1 {
		git, err := getGitClient()
		if err != nil {
			return nextTime, err
		}
		project, err := getGitProject(git)
		if err != nil {
			return nextTime, err
		}
		orderBy := "created_at"
		options := &gitlab.ListProjectIssuesOptions{
			Search:  &data.Title,
			OrderBy: &orderBy,
		}
		issues, _, err := git.Issues.ListProjectIssues(project.ID, options)
		if err != nil {
			return nextTime, err
		}
		lastIssueDate := *issues[0].CreatedAt
		lastIssueWeek := getStartOfWeek(lastIssueDate)
		currentWeek := getStartOfWeek(time.Now())
		nextIssueWeek := lastIssueWeek.AddDate(0, 0, 7*data.WeeklyRecurrence)
		daysToAdd := math.Round(nextIssueWeek.Sub(currentWeek).Hours() / 24)
		nextTime = nextTime.AddDate(0, 0, int(daysToAdd))
	}
}
