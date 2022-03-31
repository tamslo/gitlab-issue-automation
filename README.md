# GitLab Recurring Issues

A Docker image for creating recurring issues based on templates as part of a GitLab scheduled pipeline.

Forked from ⭐ [ph1ll/gitlab-recurring-issues](https://github.com/ph1ll/gitlab-recurring-issues).

## Usage

Create template issues in the `.gitlab/recurring_issue_templates/` directory as Markdown files. Template issues use YAML front matter for configuration settings. The template body is used as the issue description.

```markdown
---
title: "Biweekly reminder" # The issue title
labels: ["important", "to do"] # Optional list of labels (will be created if not present)
confidential: false
duein: "24h" # Time to due date from `crontab` as per https://pkg.go.dev/time?tab=doc#ParseDuration (e.g "30m", "1h")
crontab: "@weekly" # The recurrance schedule using crontab syntax, such as "*/30 * * * *", or a predefined value of @annually, @yearly, @monthly, @weekly, or @daily
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
