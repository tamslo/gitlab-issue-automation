package recurringIssues

import (
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	nWeeklyRecurrance "gitlab-issue-automation/n_weekly_recurrance"
	recurranceExceptions "gitlab-issue-automation/recurrance_exceptions"
	types "gitlab-issue-automation/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericaro/frontmatter"
	"github.com/gorhill/cronexpr"
)

func ProcessIssueFiles(lastRunTime time.Time) {
	err := filepath.Walk(gitlabUtils.GetRecurringIssuesPath(), processIssueFile(lastRunTime))
	if err != nil {
		log.Fatal(err)
	}
}

func parseMetadata(contents []byte) (*types.Metadata, error) {
	data := new(types.Metadata)
	err := frontmatter.Unmarshal(contents, data)
	if err != nil {
		return nil, err
	}
	monthPlaceholder := "{last_month}"
	if strings.Contains(data.Title, monthPlaceholder) {
		_, currentMonth, _ := time.Now().Date()
		lastMonth := currentMonth - 1
		data.Title = strings.ReplaceAll(data.Title, monthPlaceholder, lastMonth.String())
	}

	return data, nil
}

func getNextExecutionTime(lastTime time.Time, data *types.Metadata) time.Time {
	nextTime := data.CronExpression.Next(lastTime)
	nextTime = nWeeklyRecurrance.GetNext(nextTime, data)
	nextTime = recurranceExceptions.GetNext(nextTime, data)
	return nextTime
}

func GetRecurringIssue(path string, lastTime time.Time) (*types.Metadata, error) {
	recurringIssue := new(types.Metadata)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return recurringIssue, err
	}
	recurringIssue, err = parseMetadata(contents)
	if err != nil {
		return recurringIssue, err
	}
	cronExpression, err := cronexpr.Parse(recurringIssue.Crontab)
	if err != nil {
		return recurringIssue, err
	}
	recurringIssue.CronExpression = *cronExpression
	recurringIssue.NextTime = getNextExecutionTime(lastTime, recurringIssue)
	return recurringIssue, nil
}

func processIssueFile(lastTime time.Time) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := GetRecurringIssue(path, lastTime)
		if data.NextTime.Before(time.Now()) {
			log.Println(path, "was due", data.NextTime.Format(time.RFC3339), "- creating new issue")

			err := gitlabUtils.CreateIssue(data)
			if err != nil {
				return err
			}
		} else {
			log.Println(path, "is due", data.NextTime.Format(time.RFC3339))
		}
		return nil
	}
}
