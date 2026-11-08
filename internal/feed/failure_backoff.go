package feed

import "time"

type feedFailure struct {
	count       int
	nextAttempt time.Time
}

func (tm *TaskManager) feedCoolingDown(feedID int64, now time.Time) bool {
	tm.failureMutex.Lock()
	defer tm.failureMutex.Unlock()
	return now.Before(tm.failures[feedID].nextAttempt)
}

func (tm *TaskManager) recordFeedResult(feedID int64, success bool, now time.Time) {
	tm.failureMutex.Lock()
	defer tm.failureMutex.Unlock()
	if success {
		delete(tm.failures, feedID)
		return
	}
	if tm.failures == nil {
		tm.failures = make(map[int64]feedFailure)
	}
	failure := tm.failures[feedID]
	if failure.count < 7 {
		failure.count++
	}
	if failure.count >= 3 {
		delay := 30 * time.Minute * time.Duration(1<<(failure.count-3))
		if delay > 6*time.Hour {
			delay = 6 * time.Hour
		}
		failure.nextAttempt = now.Add(delay)
	}
	tm.failures[feedID] = failure
}
