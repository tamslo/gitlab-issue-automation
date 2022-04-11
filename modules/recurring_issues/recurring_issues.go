package recurringIssues

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericaro/frontmatter"
	"github.com/gorhill/cronexpr"
)

func GetRecurringIssuesPath() string {
	return path.Join(getCiProjectDir(), ".gitlab/recurring_issue_templates/")
}

func ParseMetadata(contents []byte) (*metadata, error) {
	data := new(metadata)
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

func ProcessIssueFile(lastTime time.Time) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		data, err := parseMetadata(contents)
		if err != nil {
			return err
		}

		cronExpression, err := cronexpr.Parse(data.Crontab)
		if err != nil {
			return err
		}

		data.NextTime, err = getNextExecutionTime(lastTime, cronExpression, data)
		if data.NextTime.Before(time.Now()) {
			log.Println(path, "was due", data.NextTime.Format(time.RFC3339), "- creating new issue")

			err := createIssue(data)
			if err != nil {
				return err
			}
		} else {
			log.Println(path, "is due", data.NextTime.Format(time.RFC3339))
		}

		return nil
	}
}

func GetNextExecutionTime(lastTime time.Time, cronExpression *cronexpr.Expression, data *metadata) (time.Time, error) {
	nextTime := cronExpression.Next(lastTime)
	nextTime = checkWeeklyRecurrance(nextTime)
	nextTime = checkRecurranceExceptions(nextTime)
	return nextTime, nil
}
