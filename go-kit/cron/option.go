package cron

import "github.com/robfig/cron/v3"

func WithSeconds() cron.Option {
	return cron.WithSeconds()
}
