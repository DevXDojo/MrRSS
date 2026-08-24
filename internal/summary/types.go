// Package summary provides text summarization using local algorithms.
// It implements TF-IDF and TextRank-based sentence scoring for extractive summarization.
package summary

// SummaryLength represents the desired length of the summary
type SummaryLength string

const (
	// Short summary with fewer sentences
	Short SummaryLength = "short"
	// Medium summary with moderate sentences
	Medium SummaryLength = "medium"
	// Long summary with more sentences
	Long SummaryLength = "long"
)

// MinContentLength is the minimum text length required for meaningful summarization
const MinContentLength = 200

// MinSentenceCount is the minimum number of sentences required for summarization
const MinSentenceCount = 3

// Target word counts for different summary lengths
// For Chinese text, each Chinese character is roughly equivalent to one English word
const (
	ShortTargetWords  = 50  // ~50 words or Chinese characters
	MediumTargetWords = 100 // ~100 words or Chinese characters
	LongTargetWords   = 150 // ~150 words or Chinese characters
)

// SummaryResult contains the generated summary and metadata
type SummaryResult struct {
	Summary       string `json:"summary"`
	Thinking      string `json:"thinking,omitempty"` // AI thinking process (optional)
	SentenceCount int    `json:"sentence_count"`
	IsTooShort    bool   `json:"is_too_short"`
}

// scoredSentence holds a sentence with its calculated score and position
type scoredSentence struct {
	text     string
	score    float64
	position int
}
