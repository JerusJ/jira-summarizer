package util

import (
	"time"

	"github.com/jerusj/jira-summarizer/pkg/constants"
)

func GetDayOfWeekOrDie(dateStr string) string {
	parsedDate, err := time.Parse(constants.DateLayoutInput, dateStr)
	if err != nil {
		panic(err)
	}
	return parsedDate.Weekday().String()
}
