package main

import (
	"database/sql"
	"sync"
)

const (
	NewSub             int = 0
	SubBeingLookedAt   int = 1
	SubGradedByStudent int = 2
	SubGradedByTA      int = 3
	SubGradedByTeacher int = 4
)

var GradingMessage = map[int]string{
	0: "Ungraded",
	1: "Correct",
	2: "Incorrect",
}

var studentWorkSnapshot = map[string]Submission{}

var sortOption = [...]string{"name", "creation_time"}

var Config *Configuration

var Database *sql.DB

var stuWrkSnapshotMutex = sync.RWMutex{}
