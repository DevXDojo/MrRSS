package statistics

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/statistics"
)

// HandleGetStatistics retrieves statistics for a specific time period
// @Summary Get statistics
// @Tags statistics
// @Param period query string false "Time period" Enums(week,month,year,all,custom) default(week)
// @Param offset query int false "Period offset for navigation (e.g., -1 for previous, 1 for next)" default(0)
// @Param start_date query string false "Start date (YYYY-MM-DD format, required for period=custom)"
// @Param end_date query string false "End date (YYYY-MM-DD format, required for period=custom)"
// @Produce json
// @Success 200 {object} StatSummary
// @Router /api/statistics [get]
func HandleGetStatistics(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	// Get period from query parameter (default to "week")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}

	// Get offset parameter (default to 0)
	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	// Get the statistics service
	statsService := h.Statistics()
	if statsService == nil {
		response.Error(w, fmt.Errorf("statistics service not available"), http.StatusInternalServerError)
		return
	}

	var summary interface{}
	var err error

	if period == "custom" {
		// Custom date range
		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")

		if startDate == "" || endDate == "" {
			response.Error(w, fmt.Errorf("start_date and end_date are required for custom period"), http.StatusBadRequest)
			return
		}

		// Validate date format
		_, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			response.Error(w, fmt.Errorf("invalid start_date format. Use YYYY-MM-DD"), http.StatusBadRequest)
			return
		}

		_, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			response.Error(w, fmt.Errorf("invalid end_date format. Use YYYY-MM-DD"), http.StatusBadRequest)
			return
		}

		summary, err = statsService.GetStatisticsForRange(startDate, endDate)
	} else {
		// Predefined period with optional offset
		validPeriods := map[string]bool{
			"week":  true,
			"month": true,
			"year":  true,
			"all":   true,
		}
		if !validPeriods[period] {
			response.Error(w, fmt.Errorf("invalid period. Must be one of: week, month, year, all, custom"), http.StatusBadRequest)
			return
		}

		summary, err = statsService.GetStatistics(statistics.StatPeriod(period), offset)
	}

	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, summary)
}

// HandleGetAllTimeStatistics retrieves all-time statistics
// @Summary Get all-time statistics
// @Tags statistics
// @Produce json
// @Success 200 {object} map[string]int
// @Router /api/statistics/all-time [get]
func HandleGetAllTimeStatistics(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	statsService := h.Statistics()
	if statsService == nil {
		response.Error(w, fmt.Errorf("statistics service not available"), http.StatusInternalServerError)
		return
	}

	stats, err := statsService.GetAllTimeStats()
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, stats)
}

// HandleGetAvailableMonths retrieves list of months with available statistics
// @Summary Get available months
// @Tags statistics
// @Produce json
// @Success 200 {array} string
// @Router /api/statistics/available-months [get]
func HandleGetAvailableMonths(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	statsService := h.Statistics()
	if statsService == nil {
		response.Error(w, fmt.Errorf("statistics service not available"), http.StatusInternalServerError)
		return
	}

	months, err := statsService.GetAvailableMonths()
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, months)
}

// HandleResetStatistics resets/deletes all statistics data
// @Summary Reset all statistics
// @Description Deletes all statistics data from the database
// @Tags statistics
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/statistics [delete]
func HandleResetStatistics(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	// Ensure it's a DELETE request
	if r.Method != http.MethodDelete {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	if h.DB == nil {
		response.Error(w, fmt.Errorf("database not available"), http.StatusInternalServerError)
		return
	}

	// Delete all statistics
	err := h.DB.ResetAllStatistics()
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]interface{}{
		"success": true,
		"message": "All statistics have been reset successfully",
	})
}
