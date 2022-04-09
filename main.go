package main

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/ericaro/frontmatter"
	"github.com/gorhill/cronexpr"
	"github.com/xanzy/go-gitlab"
)

// TODO: Refactor to not use global variables and have own modules, e.g., for basic git methods
// TODO: Test biweekly occurance, adding labels, handling exceptions

var (
	ciAPIV4URL             string = ""
	gitlabAPIToken         string = ""
	ciProjectID            string = ""
	ciProjectDir           string = ""
	ciJobName              string = ""
	shortISODateLayout     string = "2006-01-02"
	yearPlaceholder        string = "YEAR"
	issuesRelativePath     string = ".gitlab/recurring_issue_templates/"
	exceptionsRelativePath string = "./gitlab/recurring_issue_templates/recurrance_exceptions.yml"
	exceptionsExist        bool   = false
	exceptions             issueExceptions
)

type metadata struct {
	Title            string   `yaml:"title"`
	Id               string   `yaml:"id"`
	Description      string   `fm:"content" yaml:"-"`
	Confidential     bool     `yaml:"confidential"`
	Assignees        []string `yaml:"assignees,flow"`
	Labels           []string `yaml:"labels,flow"`
	DueIn            string   `yaml:"duein"`
	Crontab          string   `yaml:"crontab"`
	WeeklyRecurrence int      `yaml:"weeklyRecurrence"`
	NextTime         time.Time
}

type issueExceptions struct {
	Definitions []exceptionDefinition `yaml:"definitions"`
	Rules       []exceptionRule       `yaml:"rules"`
}

type exceptionDefinition struct {
	Id    string `yaml:"id"`
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

type exceptionRule struct {
	Issue      string   `yaml:"issue"`
	Exceptions []string `yaml:"exceptions"`
}

func parseExceptions() (issueExceptions, error) {
	_, err := os.Stat(exceptionsRelativePath)
	if err != nil {
		return exceptions, nil
	}
	exceptionsExist = true
	source, err := ioutil.ReadFile(exceptionsRelativePath)
	if err != nil {
		log.Println("No exception definition given")
		return exceptions, err
	}
	err = yaml.Unmarshal(source, &exceptions)
	if err != nil {
		return exceptions, err
	}
	return exceptions, nil
}

func areDatesEqual(aTime time.Time, anotherTime time.Time) bool {
	aYear, aMonth, aDay := aTime.Date()
	anotherYear, anotherMonth, anotherDay := anotherTime.Date()
	return aYear == anotherYear && aMonth == anotherMonth && aDay == anotherDay
}

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

func getStartOfWeek(thisTime time.Time) time.Time {
	thisWeekday := int(thisTime.Weekday())
	thisDay := time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), 0, 0, 0, 0, thisTime.Location())
	return thisDay.AddDate(0, 0, -thisWeekday)
}

func recurranceExceptionPresent(nextTime time.Time, recurringIssueId string) (bool, error) {
	exceptionPresent := false
	if recurringIssueId == "" {
		return exceptionPresent, nil
	}
	matchingExceptions := []string{}
	for _, rule := range exceptions.Rules {
		if rule.Issue == recurringIssueId {
			matchingExceptions = append(matchingExceptions,
				rule.Exceptions...)
		}
	}
	log.Println("Recurring issue ID:", recurringIssueId)
	log.Println("Matching exceptions:", matchingExceptions)
	for _, exceptionId := range matchingExceptions {
		definitionFound := false
		var exception exceptionDefinition
		for _, definition := range exceptions.Definitions {
			if exceptionId == definition.Id {
				definitionFound = true
				exception = definition
			}
		}
		if !definitionFound {
			return exceptionPresent, errors.New("Missing recurrance exception definition")
		}
		if (strings.Contains(exception.Start, yearPlaceholder) &&
			!strings.Contains(exception.End, yearPlaceholder)) ||
			(!strings.Contains(exception.Start, yearPlaceholder) &&
				strings.Contains(exception.End, yearPlaceholder)) {
			return exceptionPresent, errors.New("Please use the YEAR place holder always for both dates in the exception definition")
		}
		yearFormatLayout := "2006"
		if strings.Contains(exception.Start, yearPlaceholder) &&
			strings.Contains(exception.End, yearPlaceholder) {
			currentYear := time.Now().Format(yearFormatLayout)
			exception.Start = strings.ReplaceAll(exception.Start, yearPlaceholder, currentYear)
			exception.End = strings.ReplaceAll(exception.End, yearPlaceholder, currentYear)
			startTime, err := time.Parse(shortISODateLayout, exception.Start)
			if err != nil {
				return exceptionPresent, err
			}
			endTime, err := time.Parse(shortISODateLayout, exception.End)
			if err != nil {
				return exceptionPresent, err
			}
			if startTime.Month() > endTime.Month() {
				nextYear := time.Now().AddDate(1, 0, 0).Format(yearFormatLayout)
				exception.End = strings.ReplaceAll(exception.End, currentYear, nextYear)
			}
		}
		startTime, err := time.Parse(shortISODateLayout, exception.Start)
		if err != nil {
			return exceptionPresent, err
		}
		endTime, err := time.Parse(shortISODateLayout, exception.End)
		exceptionPresent = startTime.Before(nextTime) && endTime.After(nextTime)
		if exceptionPresent {
			log.Print("Exception present for", recurringIssueId)
			log.Print("(", exception.Id, "from", exception.Start, "to", exception.End)
			break
		}
	}
	return exceptionPresent, nil
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
	for {
		exceptionIsPresent, err := recurranceExceptionPresent(nextTime, data.Id)
		if err != nil {
			return nextTime, err
		}
		if exceptionsExist && exceptionIsPresent {
			nextTime, err := getNextExecutionTime(nextTime, cronExpression, data)
			if err != nil {
				return nextTime, err
			}
		} else {
			break
		}
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

	_, err = parseExceptions()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating recurring issues")

	err = filepath.Walk(issuesRelativePath, processIssueFile(lastRunTime))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Adapting labels")

	err = adaptLabels()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Run complete")
}
