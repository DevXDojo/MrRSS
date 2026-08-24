package statistics

import (
	"log"
	"time"

	"MrRSS/internal/database"
)

// Service handles statistics tracking and retrieval
type Service struct {
	db DB
}

// DB interface for database operations
type DB interface {
	IncrementStat(eventType string) error
	GetStatsByDateRange(startDate, endDate string) ([]database.StatRecord, error)
	GetStatsAggregated(startDate, endDate string) (map[string]int, error)
	GetStatsByDate(eventType, startDate, endDate string) (map[string]int, error)
	GetDailyStatsForPeriod(startDate, endDate string) (map[string]map[string]int, error)
	GetTotalStats() (map[string]int, error)
	GetAvailableMonths() ([]string, error)
	WaitForReady()
}

// StatPeriod represents a time period for statistics
type StatPeriod string

const (
	PeriodWeek   StatPeriod = "week"
	PeriodMonth  StatPeriod = "month"
	PeriodYear   StatPeriod = "year"
	PeriodAll    StatPeriod = "all"
	PeriodCustom StatPeriod = "custom"
)

// StatSummary represents a summary of statistics for a period
type StatSummary struct {
	Period       StatPeriod                `json:"period"`
	StartDate    string                    `json:"start_date"`
	EndDate      string                    `json:"end_date"`
	Totals       map[string]int            `json:"totals"`
	DailyData    map[string]map[string]int `json:"daily_data,omitempty"`
	CanNavigate  bool                      `json:"can_navigate"`
	HasPrevious  bool                      `json:"has_previous"`
	HasNext      bool                      `json:"has_next"`
	DisplayLabel string                    `json:"display_label"`
}

// NewService creates a new statistics service
func NewService(db DB) *Service {
	return &Service{db: db}
}

// TrackFeedRefresh tracks a feed refresh event
func (s *Service) TrackFeedRefresh() {
	if err := s.db.IncrementStat("feed_refresh"); err != nil {
		log.Printf("Error tracking feed refresh: %v", err)
	}
}

// TrackArticleRead tracks an article read event
func (s *Service) TrackArticleRead() {
	if err := s.db.IncrementStat("article_read"); err != nil {
		log.Printf("Error tracking article read: %v", err)
	}
}

// TrackAIChat tracks an AI chat event
func (s *Service) TrackAIChat() {
	if err := s.db.IncrementStat("ai_chat"); err != nil {
		log.Printf("Error tracking AI chat: %v", err)
	}
}

// TrackAISummary tracks an AI summary generation event
func (s *Service) TrackAISummary() {
	if err := s.db.IncrementStat("ai_summary"); err != nil {
		log.Printf("Error tracking AI summary: %v", err)
	}
}

// GetStatistics retrieves statistics for a specific period with optional offset
// offset allows navigating to previous/next periods (e.g., -1 for previous week, +1 for next week)
func (s *Service) GetStatistics(period StatPeriod, offset int) (*StatSummary, error) {
	now := time.Now()
	var startDate, endDate, displayLabel string
	var hasPrevious, hasNext bool

	// Apply offset to current date
	if offset != 0 {
		switch period {
		case PeriodWeek:
			// Move by weeks
			now = now.AddDate(0, 0, offset*7)
		case PeriodMonth:
			// Move by months
			now = now.AddDate(0, offset, 0)
		case PeriodYear:
			// Move by years
			now = now.AddDate(offset, 0, 0)
		}
	}

	switch period {
	case PeriodWeek:
		// Get the week (Monday to Sunday) containing the current date
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday is 0 in Go, make it 7
		}
		daysToMonday := weekday - 1
		startDate = now.AddDate(0, 0, -daysToMonday).Format("2006-01-02")
		endDate = now.AddDate(0, 0, 6-daysToMonday).Format("2006-01-02")

		// Format: "Week of Jan 1-7, 2025"
		startParsed, _ := time.Parse("2006-01-02", startDate)
		endParsed, _ := time.Parse("2006-01-02", endDate)
		displayLabel = startParsed.Format("2006-01-02") + " ~ " + endParsed.Format("2006-01-02")

		// Check navigation
		hasPrevious = true
		hasNext = true

	case PeriodMonth:
		// Get the entire month
		year, month, _ := now.Date()
		startDate = time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
		endDate = time.Date(year, month+1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1).Format("2006-01-02")

		displayLabel = now.Format("2006年01月")

		hasPrevious = true
		hasNext = true

	case PeriodYear:
		// Get the entire year
		year := now.Year()
		startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
		endDate = time.Date(year, 12, 31, 0, 0, 0, 0, time.Local).Format("2006-01-02")

		displayLabel = now.Format("2006年")

		hasPrevious = true
		hasNext = true

	case PeriodAll:
		// All time statistics
		totals, err := s.db.GetTotalStats()
		if err != nil {
			return nil, err
		}

		return &StatSummary{
			Period:       PeriodAll,
			StartDate:    "",
			EndDate:      "",
			Totals:       totals,
			CanNavigate:  false,
			HasPrevious:  false,
			HasNext:      false,
			DisplayLabel: "总计",
		}, nil

	default:
		// Default to current week
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		daysToMonday := weekday - 1
		startDate = now.AddDate(0, 0, -daysToMonday).Format("2006-01-02")
		endDate = now.AddDate(0, 0, 6-daysToMonday).Format("2006-01-02")

		startParsed, _ := time.Parse("2006-01-02", startDate)
		endParsed, _ := time.Parse("2006-01-02", endDate)
		displayLabel = startParsed.Format("2006-01-02") + " ~ " + endParsed.Format("2006-01-02")

		hasPrevious = true
		hasNext = true
	}

	// Get statistics for the calculated date range
	totals, err := s.db.GetStatsAggregated(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get daily data
	dailyData, err := s.db.GetDailyStatsForPeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &StatSummary{
		Period:       period,
		StartDate:    startDate,
		EndDate:      endDate,
		Totals:       totals,
		DailyData:    dailyData,
		CanNavigate:  true,
		HasPrevious:  hasPrevious,
		HasNext:      hasNext,
		DisplayLabel: displayLabel,
	}, nil
}

// GetStatisticsForRange retrieves statistics for a custom date range
func (s *Service) GetStatisticsForRange(startDate, endDate string) (*StatSummary, error) {
	// Get aggregated totals
	totals, err := s.db.GetStatsAggregated(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get daily data
	dailyData, err := s.db.GetDailyStatsForPeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}

	startParsed, _ := time.Parse("2006-01-02", startDate)
	endParsed, _ := time.Parse("2006-01-02", endDate)

	return &StatSummary{
		Period:       PeriodCustom,
		StartDate:    startDate,
		EndDate:      endDate,
		Totals:       totals,
		DailyData:    dailyData,
		CanNavigate:  false,
		HasPrevious:  false,
		HasNext:      false,
		DisplayLabel: startParsed.Format("2006-01-02") + " ~ " + endParsed.Format("2006-01-02"),
	}, nil
}

// GetAllTimeStats retrieves all-time statistics
func (s *Service) GetAllTimeStats() (map[string]int, error) {
	return s.db.GetTotalStats()
}

// GetAvailableMonths retrieves list of months (YYYY-MM) that have statistics data
func (s *Service) GetAvailableMonths() ([]string, error) {
	return s.db.GetAvailableMonths()
}
