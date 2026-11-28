package server

import (
	"time"
)

func NewIndexHandler() weeksInfo {
	return getWeeksInfo(time.Now().UTC(), false)
}
