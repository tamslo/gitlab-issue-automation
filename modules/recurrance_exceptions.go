package recurrance_exceptions

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

func checkRecurranceExceptions(data *metadata, currentNextTime) time.Time {
	if exceptionsExist {
		exceptionPresent, exceptionEndTime, err := recurranceExceptionPresent(nextTime, data.Id)
		if err != nil {
			return nextTime, err
		}
		if exceptionPresent {
			log.Println("Setting next time!")
			nextTime = cronExpression.Next(exceptionEndTime)
		}
	}
}

func parseExceptions() (issueExceptions, bool, error) {
	exceptionsPath := path.Join(getRecurringIssuesPath(), "recurrance_exceptions.yml")
	log.Println(exceptionsPath)
	_, err := os.Stat(exceptionsPath)
	if err != nil {
		return exceptions, exceptionsExist, nil
	}
	exceptionsExist = true
	source, err := ioutil.ReadFile(exceptionsPath)
	if err != nil {
		return exceptions, exceptionsExist, err
	}
	err = yaml.Unmarshal(source, &exceptions)
	if err != nil {
		return exceptions, exceptionsExist, err
	}
	return exceptions, exceptionsExist, nil
}

func recurranceExceptionPresent(nextTime time.Time, recurringIssueId string) (bool, time.Time, error) {
	exceptions, exceptionsExist, err = parseExceptions()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Exceptions exist:", exceptionsExist)
	log.Println("Parsed exceptions:", exceptions)
	var exceptionEndTime time.Time
	exceptionPresent := false
	if recurringIssueId == "" {
		return exceptionPresent, exceptionEndTime, nil
	}
	matchingExceptions := []string{}
	for _, rule := range exceptions.Rules {
		if rule.Issue == recurringIssueId {
			matchingExceptions = append(matchingExceptions,
				rule.Exceptions...)
		}
	}
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
			return exceptionPresent, exceptionEndTime, errors.New("Missing recurrance exception definition")
		}
		if (strings.Contains(exception.Start, yearPlaceholder) &&
			!strings.Contains(exception.End, yearPlaceholder)) ||
			(!strings.Contains(exception.Start, yearPlaceholder) &&
				strings.Contains(exception.End, yearPlaceholder)) {
			return exceptionPresent, exceptionEndTime, errors.New("Please use the YEAR place holder always for both dates in the exception definition")
		}
		yearFormatLayout := "2006"
		if strings.Contains(exception.Start, yearPlaceholder) &&
			strings.Contains(exception.End, yearPlaceholder) {
			currentYear := time.Now().Format(yearFormatLayout)
			exception.Start = strings.ReplaceAll(exception.Start, yearPlaceholder, currentYear)
			exception.End = strings.ReplaceAll(exception.End, yearPlaceholder, currentYear)
			startTime, err := time.Parse(shortISODateLayout, exception.Start)
			if err != nil {
				return exceptionPresent, exceptionEndTime, err
			}
			endTime, err := time.Parse(shortISODateLayout, exception.End)
			if err != nil {
				return exceptionPresent, exceptionEndTime, err
			}
			if startTime.Month() > endTime.Month() {
				nextYear := time.Now().AddDate(1, 0, 0).Format(yearFormatLayout)
				exception.End = strings.ReplaceAll(exception.End, currentYear, nextYear)
			}
		}
		startTime, err := time.Parse(shortISODateLayout, exception.Start)
		if err != nil {
			return exceptionPresent, exceptionEndTime, err
		}
		endTime, err := time.Parse(shortISODateLayout, exception.End)
		exceptionPresent = startTime.Before(nextTime) && endTime.After(nextTime)
		if exceptionPresent {
			log.Println("Applying exception", exception.Id, "for", recurringIssueId, "from", exception.Start, "to", exception.End)
			exceptionEndTime = endTime
			break
		}
	}
	return exceptionPresent, exceptionEndTime, nil
}
