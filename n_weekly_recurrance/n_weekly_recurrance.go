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
		lastIssueDate := *issues[0].CreatedAt
		lastIssueWeek := dateUtils.GetStartOfWeek(lastIssueDate)
		currentWeek := dateUtils.GetStartOfWeek(time.Now())
		nextIssueWeek := lastIssueWeek.AddDate(0, 0, 7*data.WeeklyRecurrence)
		daysToAdd := math.Round(nextIssueWeek.Sub(currentWeek).Hours() / 24)
		if verbose {
			dueInWeeks := math.Round(math.Abs(daysToAdd / 7))
			log.Println("-- Next", data.WeeklyRecurrence, "weekly occurrence for", data.Title, "in", dueInWeeks, "week(s)")
		}
		nextTime = nextTime.AddDate(0, 0, int(daysToAdd))
		log.Println("-- Degugging start")
		log.Println("--- Ordered issues")
		log.Printf("%+v\n", issues)
		log.Println("--- Last date")
		log.Println(lastIssueDate)
		log.Println("--- Last week")
		log.Println(lastIssueWeek)
		log.Println("--- Current week")
		log.Println(currentWeek)
		log.Println("--- Next week")
		log.Println(nextIssueWeek)
		log.Println("---Days to add")
		log.Println(daysToAdd)
		log.Println("--- Next time")
		log.Println(nextTime)
		log.Println("-- Degugging end")
	}
	return nextTime
}
