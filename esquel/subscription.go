package esquel

import "time"

type subscription func(query string, duration time.Duration)
