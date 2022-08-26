package recurringIssues

import (
	"gitlab-issue-automation/constants"
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	nWeeklyRecurrance "gitlab-issue-automation/n_weekly_recurrance"
	placeholders "gitlab-issue-automation/placeholders"
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
	return data, nil
}

func getNextExecutionTime(lastTime time.Time, data *types.Metadata, verbose bool) time.Time {
	nextTime := data.CronExpression.Next(lastTime)
	nextTime = nWeeklyRecurrance.GetNext(nextTime, data, verbose)
	nextTime = recurranceExceptions.GetNext(nextTime, data, verbose)
	return nextTime
}

func GetRecurringIssue(path string, lastTime time.Time, verbose bool) (*types.Metadata, error) {
	recurringIssue := new(types.Metadata)
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return recurringIssue, err
	}
	recurringIssue, err = parseMetadata(contents)
	if err != nil {
		return recurringIssue, err
	}
	if strings.HasSuffix(path, constants.VacationTemplateName) {
		log.Println("- TODO: Implement vacation issue creation")
		// if recurranceExceptions.IsVacationUpcoming() {}
	} else {
		cronExpression, err := cronexpr.Parse(recurringIssue.Crontab)
		if err != nil {
			return recurringIssue, err
		}
		recurringIssue.CronExpression = *cronExpression
		recurringIssue.NextTime = getNextExecutionTime(lastTime, recurringIssue, verbose)
	}
	recurringIssue = placeholders.ApplyPlaceholders(recurringIssue)
	return recurringIssue, nil
}

func processIssueFile(lastTime time.Time) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		log.Println("- Checking", path)
		verbose := true
		data, err := GetRecurringIssue(path, lastTime, verbose)
		if err != nil {
			return err
		}
		if data.NextTime.Before(time.Now()) {
			log.Println("--", info.Name(), "was due", data.NextTime.Format(time.RFC3339), "- creating new issue")

			err := gitlabUtils.CreateIssue(data)
			if err != nil {
				return err
			}
		} else {
			log.Println("--", info.Name(), "will be due", data.NextTime.Format(time.RFC3339))
		}
		return nil
	}
}
