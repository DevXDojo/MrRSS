//go:build !server

package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"reflect"
	"testing"
)

type recordingWindow struct {
	application.Window // Unexpected calls such as Restore/SetSize fail the test.
	calls              []string
}

func (w *recordingWindow) Show() application.Window { w.calls = append(w.calls, "show"); return w }
func (w *recordingWindow) UnMinimise()              { w.calls = append(w.calls, "unminimize") }
func (w *recordingWindow) Maximise() application.Window {
	w.calls = append(w.calls, "maximize")
	return w
}
func (w *recordingWindow) Focus() { w.calls = append(w.calls, "focus") }

func TestShowExistingWindowPreservesPlacement(t *testing.T) {
	for _, maximize := range []bool{false, true} {
		window := &recordingWindow{}
		showExistingWindow(window, maximize)
		want := []string{"show", "unminimize", "focus"}
		if maximize {
			want = []string{"show", "unminimize", "maximize", "focus"}
		}
		if !reflect.DeepEqual(window.calls, want) {
			t.Fatalf("maximize=%v calls=%v", maximize, window.calls)
		}
	}
}
