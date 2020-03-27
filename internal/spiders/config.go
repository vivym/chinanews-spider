package spiders

import "time"

var (
	startDate = time.Now().Truncate(24 * time.Hour).Add(-8 * time.Hour)
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.130 Safari/537.36"
)

type Config struct {
	Concurrency int
	Delay       int
	Tasks       map[string][]Task
}

type Task struct {
	Province string
	BaseURL  string
	Selector *Selector
	Skip     bool
	SubTasks []SubTask
}

type SubTask struct {
	Tag      string
	URL      string
	Selector *Selector
}

type Selector struct {
	URL     string
	Title   string
	Src     string
	Content string
}
