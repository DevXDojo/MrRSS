package translation

import "testing"

// Chinese titles are checked even when translation is skipped. Keep this hot
// path benchmarkable without contacting a translation service.
func BenchmarkChineseVariant(b *testing.B) {
	// Exclude the one-time dictionary initialization from the steady-state cost.
	if got := detectChineseVariant("这是一篇关于技术和编程的测试文章。"); got != "zh" {
		b.Fatal(got)
	}
	b.ReportAllocs()
	for b.Loop() {
		if got := detectChineseVariant("这是一篇关于技术和编程的测试文章。"); got != "zh" {
			b.Fatal(got)
		}
	}
}
