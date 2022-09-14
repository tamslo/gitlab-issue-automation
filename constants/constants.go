package constants

const IssueTemplatePath = ".gitlab/recurring_issue_templates/"
const StandupIssueTemplateName = "prepare-standup.md" // for this template notes will be created

// Vacation issue definitions

const VacationTemplateName = "vacation.md"
const VacationExceptionPrefix = "vacation-"

// Label definitions

const ThisWeekLabel = "🗓 This week"
const TodayLabel = "☀️ Today"
const InProgressLabel = "🏃‍♀️ In progress"
const WaitingLabel = "⏳ Waiting"
const InOfficeLabel = "🏢 In office"
const RecurringLabel = "🔁 Recurring"
const NextActionsLabel = "⏭ Next actions"
const SomewhenLabel = "🔮 Somewhen"
const TestLabel = "🧪 Test"

var ProgressLabels = []string{InProgressLabel, InOfficeLabel, WaitingLabel}
var StatusLabels = []string{ThisWeekLabel, TodayLabel, InProgressLabel, WaitingLabel, InOfficeLabel}
var NonProjectLabels = []string{ThisWeekLabel, TodayLabel, InProgressLabel, WaitingLabel, InOfficeLabel, RecurringLabel, NextActionsLabel, SomewhenLabel, TestLabel}
