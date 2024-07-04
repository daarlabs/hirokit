package convertx

import (
	"time"
	
	"github.com/daarlabs/hirokit/constant/dataFormat"
)

func ConvertStringToTime(value string) time.Time {
	t, err := time.Parse(dataFormat.UTCTime, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

func ConvertTimeToString(t time.Time) string {
	return t.UTC().String()
}
