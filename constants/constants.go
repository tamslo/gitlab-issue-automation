package constants

const IssueTemplatePath = ".gitlab/recurring_issue_templates/"
const StandupIssueTemplateName = "prepare-standup.md" // for this template notes will be created

// Label definitions

const ThisWeekLabel = "🗓 This week"
const TodayLabel = "☀️ Today"
const InProgressLabel = "🏃‍♀️ In progress"
const WaitingLabel = "⏳ Waiting"
const InOfficeLabel = "🏢 In office"
const RecurringLabel = "🔁 Recurring"
const TestLabel = "🧪 Test"

var ProgressLabels = []string{InProgressLabel, InOfficeLabel, WaitingLabel}
var NonProjectLabels = []string{ThisWeekLabel, TodayLabel, InProgressLabel, WaitingLabel, InOfficeLabel, RecurringLabel, TestLabel}
