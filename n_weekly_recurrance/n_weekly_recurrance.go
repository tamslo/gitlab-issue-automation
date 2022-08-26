package weeklyRecurrance

import (
	dateUtils "gitlab-issue-automation/date_utils"
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	types "gitlab-issue-automation/types"
	"log"
	"math"
	"time"

	"github.com/xanzy/go-gitlab"
)

func GetNext(nextTime time.Time, data *types.Metadata, verbose bool) time.Time {
	if data.WeeklyRecurrence > 1 {
		git := gitlabUtils.GetGitClient()
		project := gitlabUtils.GetGitProject()
		orderBy := "created_at"
		options := &gitlab.ListProjectIssuesOptions{
			Search:  &data.Title,
			OrderBy: &orderBy,
		}
		issues, _, err := git.Issues.ListProjectIssues(project.ID, options)
		if err != nil {
			log.Fatal(err)
		}
		lastCreationDate := *issues[0].CreatedAt
		lastCreationWeek := dateUtils.GetStartOfWeek(lastCreationDate)
		nextSingleExecutionWeek := dateUtils.GetStartOfWeek(nextTime)
		nextNthExecutionWeek := lastCreationWeek
		for {
			nextNthExecutionWeek = nextNthExecutionWeek.AddDate(0, 0, 7*data.WeeklyRecurrence)
			if dateUtils.AreDatesEqual(nextNthExecutionWeek, nextSingleExecutionWeek) || nextNthExecutionWeek.After(nextSingleExecutionWeek) {
				break
			}
		}
		daysToAdd := math.Round(nextNthExecutionWeek.Sub(nextSingleExecutionWeek).Hours() / 24)
		if verbose {
			weeksToAdd := math.Round(math.Abs(daysToAdd / 7))
			log.Println("-- Next", data.WeeklyRecurrence, "weekly occurrence for", data.Title, "in plus", weeksToAdd, "week(s)")
		}
		nextTime = nextTime.AddDate(0, 0, int(daysToAdd))
	}
	return nextTime
}
