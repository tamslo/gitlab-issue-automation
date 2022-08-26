package recurrance_exceptions

import (
	"errors"
	"fmt"
	dateUtils "gitlab-issue-automation/date_utils"
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	types "gitlab-issue-automation/types"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

func GetNext(nextTime time.Time, data *types.Metadata, verbose bool) time.Time {
	if data.Id != "" && exceptionsExist() {
		exceptions := parseExceptions()
		matchingExceptions := getExceptionIdsForIssue(exceptions, data.Id)
		for _, exceptionId := range matchingExceptions {
			exceptionDefinition := getExceptionDefinition(exceptions.Definitions, exceptionId)
			startTime, err := time.Parse(dateUtils.ShortISODateLayout, exceptionDefinition.Start)
			if err != nil {
				log.Fatal(err)
			}
			endTime, err := time.Parse(dateUtils.ShortISODateLayout, exceptionDefinition.End)
			if err != nil {
				log.Fatal(err)
			}
			// Need to additioanlly check if dates are equal if day but not time fulfill the before or after condition
			exceptionApplies := (endTime.After(time.Now()) || dateUtils.AreDatesEqual(endTime, time.Now())) &&
				(startTime.Before(nextTime) || dateUtils.AreDatesEqual(startTime, nextTime)) &&
				(endTime.After(nextTime) || dateUtils.AreDatesEqual(endTime, nextTime))
			if exceptionApplies {
				nextTime = data.CronExpression.Next(endTime.AddDate(0, 0, 1))
				if verbose {
					log.Println("-- Applying exception", exceptionDefinition.Id, "for", data.Id, "from", exceptionDefinition.Start, "to", exceptionDefinition.End)
					log.Println("-- Setting earliest execution date after exception (ignoring n-weekly recurrances for now)")
				}
				break
			}
		}
	}
	return nextTime
}

func getExceptionIdsForIssue(exceptions types.RecurranceExceptions, issueId string) []string {
	matchingExceptions := []string{}
	for _, rule := range exceptions.Rules {
		if rule.Issue == issueId {
			matchingExceptions = append(matchingExceptions, rule.Exceptions...)
		}
	}
	return matchingExceptions
}

func getExceptionDefinition(exceptionDefinitions []types.ExceptionDefinition, exceptionId string) types.ExceptionDefinition {
	definitionFound := false
	var exceptionDefinition types.ExceptionDefinition
	for _, definition := range exceptionDefinitions {
		if exceptionId == definition.Id {
			definitionFound = true
			exceptionDefinition = definition
		}
	}
	if !definitionFound {
		log.Fatal(fmt.Errorf("unknown exception definition %s", exceptionId))
	}
	return fillInYearPlaceholdes(exceptionDefinition)
}

func fillInYearPlaceholdes(exceptionDefinition types.ExceptionDefinition) types.ExceptionDefinition {
	const YearPlaceholder = "YEAR"
	if (strings.Contains(exceptionDefinition.Start, YearPlaceholder) &&
		!strings.Contains(exceptionDefinition.End, YearPlaceholder)) ||
		(!strings.Contains(exceptionDefinition.Start, YearPlaceholder) &&
			strings.Contains(exceptionDefinition.End, YearPlaceholder)) {
		log.Fatal(errors.New("please use the YEAR place holder always for both dates in the exception definition"))
	}
	if strings.Contains(exceptionDefinition.Start, YearPlaceholder) &&
		strings.Contains(exceptionDefinition.End, YearPlaceholder) {
		currentYear := time.Now().Format(dateUtils.YearDateLayout)
		exceptionDefinition.Start = strings.ReplaceAll(exceptionDefinition.Start, YearPlaceholder, currentYear)
		exceptionDefinition.End = strings.ReplaceAll(exceptionDefinition.End, YearPlaceholder, currentYear)
		startTime, err := time.Parse(dateUtils.ShortISODateLayout, exceptionDefinition.Start)
		if err != nil {
			log.Fatal(err)
		}
		endTime, err := time.Parse(dateUtils.ShortISODateLayout, exceptionDefinition.End)
		if err != nil {
			log.Fatal(err)
		}
		if startTime.Month() > endTime.Month() {
			nextYear := time.Now().AddDate(1, 0, 0).Format(dateUtils.YearDateLayout)
			exceptionDefinition.End = strings.ReplaceAll(exceptionDefinition.End, currentYear, nextYear)
		}
	}
	return exceptionDefinition
}

func parseExceptions() types.RecurranceExceptions {
	exceptionsPath := getExceptionsPath()
	exceptions := types.RecurranceExceptions{}
	source, err := ioutil.ReadFile(exceptionsPath)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(source, &exceptions)
	if err != nil {
		log.Fatal(err)
	}
	return exceptions
}

func exceptionsExist() bool {
	exceptionsPath := getExceptionsPath()
	_, err := os.Stat(exceptionsPath)
	return err == nil
}

func getExceptionsPath() string {
	return path.Join(gitlabUtils.GetRecurringIssuesPath(), "recurrance_exceptions.yml")
}
