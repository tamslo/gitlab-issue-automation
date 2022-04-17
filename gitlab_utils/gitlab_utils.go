package gitlabUtils

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	types "gitlab-issue-automation/types"

	"github.com/xanzy/go-gitlab"
)

type envVariableParameters struct {
	Name                  string
	ErrorMessageOverwrite string
}

func GetEnvVariable(parameters *envVariableParameters) string {
	envVariable := os.Getenv(parameters.Name)
	if envVariable == "" {
		errorMessage := "This tool must be ran as part of a GitLab pipeline."
		if parameters.ErrorMessageOverwrite != "" {
			errorMessage = parameters.ErrorMessageOverwrite
		}
		log.Fatal("Environment variable", envVariable, "not found.", errorMessage)
	}
	return envVariable
}

func GetGitlabAPIToken() string {
	return GetEnvVariable(&envVariableParameters{
		Name:                  "GITLAB_API_TOKEN",
		ErrorMessageOverwrite: "Ensure this is set under the project CI/CD settings.",
	})
}

func GetCiProjectId() string {
	return GetEnvVariable(&envVariableParameters{Name: "CI_PROJECT_ID"})
}

func GetCiAPIV4URL() string {
	return GetEnvVariable(&envVariableParameters{Name: "CI_API_V4_URL"})
}

func GetCiProjectDir() string {
	return GetEnvVariable(&envVariableParameters{Name: "CI_PROJECT_DIR"})
}

func GetCiJobName() string {
	return GetEnvVariable(&envVariableParameters{Name: "CI_JOB_NAME"})
}

func GetGitClient() *gitlab.Client {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{
		Transport: transCfg,
	}
	git, err := gitlab.NewClient(GetGitlabAPIToken(), gitlab.WithBaseURL(GetCiAPIV4URL()), gitlab.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatal(err)
	}
	return git
}

func GetGitProject() *gitlab.Project {
	git := GetGitClient()
	project, _, err := git.Projects.GetProject(GetCiProjectId(), nil)
	if err != nil {
		log.Fatal(err)
	}
	return project
}

func GetRecurringIssuesPath() string {
	return path.Join(GetCiProjectDir(), ".gitlab/recurring_issue_templates/")
}

func GetLastRunTime() time.Time {
	git := GetGitClient()
	lastRunTime := time.Unix(0, 0)
	options := &gitlab.ListProjectPipelinesOptions{
		Scope:   gitlab.String("finished"),
		Status:  gitlab.BuildState(gitlab.Success),
		OrderBy: gitlab.String("updated_at"),
	}
	ciProjectID := GetCiProjectId()
	ciJobName := GetCiJobName()
	pipelineInfos, _, err := git.Pipelines.ListProjectPipelines(ciProjectID, options)
	if err != nil {
		log.Fatal(err)
	}
	for _, pipelineInfo := range pipelineInfos {
		jobs, _, err := git.Jobs.ListPipelineJobs(ciProjectID, pipelineInfo.ID, nil)
		if err != nil {
			log.Fatal(err)
		}
		for _, job := range jobs {
			if job.Name == ciJobName {
				lastRunTime = *job.FinishedAt
				return lastRunTime
			}
		}
	}
	return lastRunTime
}

func GetSortedProjectIssues(orderBy string, sortOrder string) []*gitlab.Issue {
	git := GetGitClient()
	project := GetGitProject()
	issueState := "opened"
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
			log.Fatal(err)
		}
		issues = append(issues, pageIssues...)
		if len(pageIssues) < perPage {
			lastPageReached = true
		} else {
			page++
		}
	}
	return issues
}

func CreateIssue(data *types.Metadata) error {
	git := GetGitClient()
	project := GetGitProject()
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
	_, _, err := git.Issues.CreateIssue(project.ID, options)
	if err != nil {
		return err
	}
	return nil
}

func UpdateIssue(issueId int, options *gitlab.UpdateIssueOptions) *gitlab.Issue {
	git := GetGitClient()
	project := GetGitProject()
	updatedIssue, _, err := git.Issues.UpdateIssue(project.ID, issueId, options)
	if err != nil {
		log.Fatal(err)
	}
	return updatedIssue
}
