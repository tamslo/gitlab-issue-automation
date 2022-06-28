package issueTypes

import (
	"time"

	"github.com/gorhill/cronexpr"
)

type Metadata struct {
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
	CronExpression   cronexpr.Expression
}

type RecurranceExceptions struct {
	Definitions []ExceptionDefinition `yaml:"definitions"`
	Rules       []ExceptionRule       `yaml:"rules"`
}

type ExceptionDefinition struct {
	Id    string `yaml:"id"`
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

type ExceptionRule struct {
	Issue      string   `yaml:"issue"`
	Exceptions []string `yaml:"exceptions"`
}

type WikiMetadata struct {
	Title string
	Slug  string
}
