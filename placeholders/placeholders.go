package placeholders

import (
	dateUtils "gitlab-issue-automation/date_utils"
	gitlabUtils "gitlab-issue-automation/gitlab_utils"
	types "gitlab-issue-automation/types"
	"strings"
	"time"
)

const lastMonthPlaceholder = "{last_month}"

func getLastMonth(data *types.Metadata) string {
	_, currentMonth, _ := time.Now().Date()
	lastMonth := currentMonth - 1
	lastMonthString := lastMonth.String()
	return lastMonthString
}

const dateEnDashPlaceholder = "{due_date_en_dash}"

func getEnDashDate(data *types.Metadata) string {
	issueDue := gitlabUtils.GetIssueDueDate(data)
	enDashDate := dateUtils.GetEnDashDate(issueDue)
	return enDashDate
}

var placeholders = map[string]func(*types.Metadata) string{
	lastMonthPlaceholder:  getLastMonth,
	dateEnDashPlaceholder: getEnDashDate,
}

func applyPlaceholder(data *types.Metadata, placeholder string, replacement string) *types.Metadata {
	if strings.Contains(data.Title, placeholder) {
		data.Title = strings.ReplaceAll(data.Title, placeholder, replacement)
	}
	if strings.Contains(data.Description, placeholder) {
		data.Description = strings.ReplaceAll(data.Description, placeholder, replacement)
	}
	return data
}

func ApplyPlaceholders(data *types.Metadata) *types.Metadata {
	for placeholder, getPlaceholderValue := range placeholders {
		data = applyPlaceholder(data, placeholder, getPlaceholderValue(data))
	}
	return data
}
