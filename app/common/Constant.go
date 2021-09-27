package common

import "strings"

const (
	JobSaveDir   string = "/cron/jobs/"
	JobKillDir   string = "/cron/kill/"
	JobLockDir   string = "/cron/lock/"
	JobWorkerDir string = "/cron/worker/"
)
const (
	TimeLayout string = "2006-01-02 15:04:05"
)

func ExtractKeyName(key, prefix string) string {
	return strings.TrimPrefix(key, prefix)
}
