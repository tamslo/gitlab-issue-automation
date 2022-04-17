# GitLab Recurring Issues

A Docker image for creating recurring issues based on templates as part of a GitLab scheduled pipeline.

Forked from ⭐ [ph1ll/gitlab-recurring-issues](https://github.com/ph1ll/gitlab-recurring-issues) and adapted for my personal use.

The Docker image is available on [DockerHub](https://hub.docker.com/repository/docker/tamslo/gitlab-issue-automation).

## Usage

Create template issues in the `.gitlab/recurring_issue_templates/` directory as Markdown files. Template issues use YAML front matter for configuration settings. The template body is used as the issue description.

```markdown
---
title: "Biweekly reminder" # The issue title
labels: ["important", "to do"] # Optional; list of labels (will be created if not present)
confidential: false # Optional; defines visibility of issue (default for bool in Go is false)
duein: "24h" # Optional; time to due date from `crontab` as per https://pkg.go.dev/time?tab=doc#ParseDuration (e.g "30m", "1h")
crontab: "@weekly" # The recurrance schedule for issue creation using crontab syntax
weeklyRecurrence: 2 # Optional; if stated, the `crontab` condition will only be applied to every n-th week, based on titles of present issues
---
(**You need to give a description, otherwise parsing will fail!**)

This is your biweekly reminder to perform the following actions:

* [ ] Action 1
* [ ] Action 2
```

Create a pipeline in the `.gitlab-ci.yml` file:

```yaml
recurring issues:
  image: tamslo/gitlab-recurring-issues
  script: gitlab-recurring-issues
  only: 
    - schedules
```

Create project CI/CD variables:

| Name | Value |
| ---- | ----- |
| GITLAB_API_TOKEN | The API access token for the user account that will create the issues (see: https://docs.gitlab.com/ce/user/profile/personal_access_tokens.html) | 

Finally, create a new schedule under the project CI/CD options, ensuring that the pipeline runs at least as often as your most frequent job.

### Adding Recurrance Exceptions

To add exceptions to recurrances, create a file named `recurrance_exceptions.yml` in the templates folder. Note that exception dates are applied to the creation date given in `crontab`, not the due date.

It can contain exception definitions and rules that map issues by their IDs (need to be given in the issue template) to exception definitions.

Start and end dates are given in the format `YYYY-MM-DD`. If an exception occurs every year, the placeholder `YEAR` can be given (needs to be set for both `start` and `end`). 

```
definitions:
  -
    id: "christmas-break"
    start: "YEAR-12-24"
    end: "YEAR-01-01"
  -
    id: "vacation"
    start: "2022-05-13"
    end: "2022-05-20"
  -
    id: "no-meeting"
    start: "2022-04-20"
    end: "2022-04-20"
rules:
  -
    issue: "weekly-meeting"
    exceptions: ["christmas-break", "vacation", "no-meeting"]
```

### Automatically Moving Issues on Board

The script also checks whether labels for custom issue management on a board view exist (see `adaptaLabels`).

If an issue is due, the `TodayLabel` or `ThisWeekLabel` will be added if it is not present and no `OtherLabels` exist that indicate that the issue is in progress. If the `TodayLabel` is added and the `ThisWeekLabel` present, the latter will be removed.