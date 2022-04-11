package dateUtils

import "time"

shortISODateLayout := "2006-01-02"

func getStartOfWeek(thisTime time.Time) time.Time {
	thisWeekday := int(thisTime.Weekday())
	thisDay := time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), 0, 0, 0, 0, thisTime.Location())
	return thisDay.AddDate(0, 0, -thisWeekday)
}

func areDatesEqual(aTime time.Time, anotherTime time.Time) bool {
	aYear, aMonth, aDay := aTime.Date()
	anotherYear, anotherMonth, anotherDay := anotherTime.Date()
	return aYear == anotherYear && aMonth == anotherMonth && aDay == anotherDay
}
