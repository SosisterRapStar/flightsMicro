package v1

import "time"

const (
	dateLayout = "2006-01-02"
)

func parseDate(value string) (time.Time, error) {
	return time.Parse(dateLayout, value)
}

func parseDateTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}
