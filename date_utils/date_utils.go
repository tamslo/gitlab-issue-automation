package dateUtils

import (
	"strings"
	"time"
)

const ShortISODateLayout = "2006-01-02"

const YearDateLayout = "2006"

const enDash = "–"
const dash = "-"

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

func EscapeDashes(text string) string {
	return strings.ReplaceAll(text, dash, enDash)
}

func UnescapeDashes(text string) string {
	return strings.ReplaceAll(text, enDash, dash)
}

func GetEnDashDate(date time.Time) string {
	return EscapeDashes(date.Format(ShortISODateLayout))
}
