package dateUtils

import "time"

const ShortISODateLayout = "2006-01-02"

const YearDateLayout = "2006"

func GetStartOfWeek(thisTime time.Time) time.Time {
	thisWeekday := int(thisTime.Weekday())
	thisDay := time.Date(thisTime.Year(), thisTime.Month(), thisTime.Day(), 0, 0, 0, 0, thisTime.Location())
	return thisDay.AddDate(0, 0, -thisWeekday)
}

func AreDatesEqual(aTime time.Time, anotherTime time.Time) bool {
	aYear, aMonth, aDay := aTime.Date()
	anotherYear, anotherMonth, anotherDay := anotherTime.Date()
	return aYear == anotherYear && aMonth == anotherMonth && aDay == anotherDay
}
