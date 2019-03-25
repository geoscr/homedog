package ORM

import "time"

type Post struct {
	Id        int64
	Recip     string
	Counter   int64
	Source    string
	Title     string
	Body      string
	Url       string
	Timestamp time.Time
}
