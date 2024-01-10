package main

import (
	"sync"
	"time"
)

// This hold student's latest submission time
var studentLastSubmission = map[int]time.Time{}

var studentWorkSnapshot = map[string]Submission{}

var stuWrkSnapshotMutex = sync.RWMutex{}
