package main

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericaro/frontmatter"
	"github.com/gorhill/cronexpr"
	"github.com/xanzy/go-gitlab"
)

var (
	ciAPIV4URL         string = ""
	gitlabAPIToken     string = ""
	ciProjectID        string = ""
	ciProjectDir       string = ""
	ciJobName          string = ""
	issuesRelativePath string = ".gitlab/recurring_issue_templates/"
)

type metadata struct {
	Title            string   `yaml:"title"`
	Description      string   `fm:"content" yaml:"-"`
	Confidential     bool     `yaml:"confidential"`
	Assignees        []string `yaml:"assignees,flow"`
	Labels           []string `yaml:"labels,flow"`
	DueIn            string   `yaml:"duein"`
	Crontab          string   `yaml:"crontab"`
	WeeklyRecurrence int      `yaml:"weeklyRecurrence"`
	NextTime         time.Time
}

func getStartOfWeek(thisTime time.Time) time.Time {
	thisWeekday := int(thisTime.Weekday())
	thisDay := time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), 0, 0, 0, 0, thisTime.Location())
	return thisDay.AddDate(0, 0, -thisWeekday)
}

func getNextExecutionTime(lastTime time.Time, cronExpression *cronexpr.Expression, data *metadata) (time.Time, error) {
	nextTime := cronExpression.Next(lastTime)
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
	return nextTime, nil
}

func processIssueFile(lastTime time.Time) filepath.WalkFunc {
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

func parseMetadata(contents []byte) (*metadata, error) {
	data := new(metadata)
	err := frontmatter.Unmarshal(contents, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getGitClient() (*gitlab.Client, error) {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{
		Transport: transCfg,
	}
	return gitlab.NewClient(gitlabAPIToken, gitlab.WithBaseURL(ciAPIV4URL), gitlab.WithHTTPClient(httpClient))
}

func getGitProject(git *gitlab.Client) (*gitlab.Project, error) {
	project, _, err := git.Projects.GetProject(ciProjectID, nil)
	return project, err
}

func createIssue(data *metadata) error {
	git, err := getGitClient()
	if err != nil {
		return err
	}

	project, err := getGitProject(git)
	if err != nil {
		return err
	}

	monthPlaceholder := "{last_month}"
	if strings.Contains(data.Title, monthPlaceholder) {
		_, currentMonth, _ := time.Now().Date()
		lastMonth := currentMonth - 1
		data.Title = strings.ReplaceAll(data.Title, monthPlaceholder, lastMonth.String())
	}

	options := &gitlab.CreateIssueOptions{
		Title:        gitlab.String(data.Title),
		Description:  gitlab.String(data.Description),
		Confidential: &data.Confidential,
		CreatedAt:    &data.NextTime,
		Labels:       &gitlab.Labels{strings.Join(data.Labels, ",")},
	}

	if data.DueIn != "" {
		duration, err := time.ParseDuration(data.DueIn)
		if err != nil {
			return err
		}

		dueDate := gitlab.ISOTime(data.NextTime.Add(duration))

		options.DueDate = &dueDate
	}

	_, _, err = git.Issues.CreateIssue(project.ID, options)
	if err != nil {
		return err
	}

	return nil
}

func getLastRunTime() (time.Time, error) {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{
		Transport: transCfg,
	}

	git, err := gitlab.NewClient(gitlabAPIToken, gitlab.WithBaseURL(ciAPIV4URL), gitlab.WithHTTPClient(httpClient))
	if err != nil {
		return time.Unix(0, 0), err
	}

	options := &gitlab.ListProjectPipelinesOptions{
		Scope:   gitlab.String("finished"),
		Status:  gitlab.BuildState(gitlab.Success),
		OrderBy: gitlab.String("updated_at"),
	}

	pipelineInfos, _, err := git.Pipelines.ListProjectPipelines(ciProjectID, options)
	if err != nil {
		return time.Unix(0, 0), err
	}

	for _, pipelineInfo := range pipelineInfos {
		jobs, _, err := git.Jobs.ListPipelineJobs(ciProjectID, pipelineInfo.ID, nil)
		if err != nil {
			return time.Unix(0, 0), err
		}

		for _, job := range jobs {
			if job.Name == ciJobName {
				return *job.FinishedAt, nil
			}
		}
	}

	return time.Unix(0, 0), nil
}

func main() {
	gitlabAPIToken = os.Getenv("GITLAB_API_TOKEN")
	if gitlabAPIToken == "" {
		log.Fatal("Environment variable 'GITLAB_API_TOKEN' not found. Ensure this is set under the project CI/CD settings.")
	}

	ciAPIV4URL = os.Getenv("CI_API_V4_URL")
	if ciAPIV4URL == "" {
		log.Fatal("Environment variable 'CI_API_V4_URL' not found. This tool must be ran as part of a GitLab pipeline.")
	}

	ciProjectID = os.Getenv("CI_PROJECT_ID")
	if gitlabAPIToken == "" {
		log.Fatal("Environment variable 'CI_PROJECT_ID' not found. This tool must be ran as part of a GitLab pipeline.")
	}

	ciProjectDir = os.Getenv("CI_PROJECT_DIR")
	if gitlabAPIToken == "" {
		log.Fatal("Environment variable 'CI_PROJECT_DIR' not found. This tool must be ran as part of a GitLab pipeline.")
	}

	ciJobName = os.Getenv("CI_JOB_NAME")
	if gitlabAPIToken == "" {
		log.Fatal("Environment variable 'CI_JOB_NAME' not found. This tool must be ran as part of a GitLab pipeline.")
	}

	issuesRelativePath = path.Join(ciProjectDir, issuesRelativePath)

	lastRunTime, err := getLastRunTime()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Last run:", lastRunTime.Format(time.RFC3339))

	err = filepath.Walk(issuesRelativePath, processIssueFile(lastRunTime))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Run complete")
}
