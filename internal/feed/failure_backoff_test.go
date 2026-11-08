package feed

import (
	"testing"
	"time"
)

func TestScheduledFeedFailureBackoff(t *testing.T) {
	tm := &TaskManager{}
	now := time.Now()
	for i := 0; i < 2; i++ {
		tm.recordFeedResult(1, false, now)
	}
	if tm.feedCoolingDown(1, now) {
		t.Fatal("backoff started before three failures")
	}
	tm.recordFeedResult(1, false, now)
	if !tm.feedCoolingDown(1, now.Add(29*time.Minute)) || tm.feedCoolingDown(1, now.Add(30*time.Minute)) {
		t.Fatal("incorrect initial delay")
	}
	for i := 0; i < 20; i++ {
		tm.recordFeedResult(1, false, now)
	}
	if tm.feedCoolingDown(1, now.Add(6*time.Hour)) {
		t.Fatal("backoff exceeds six hours")
	}
	tm.recordFeedResult(1, true, now)
	if tm.feedCoolingDown(1, now) {
		t.Fatal("success did not clear backoff")
	}
}
