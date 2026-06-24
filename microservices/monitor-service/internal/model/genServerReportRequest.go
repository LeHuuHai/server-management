package model

import "time"

type GenServerReportRequest struct {
	From      time.Time
	To        time.Time
	Receivers []string
}
