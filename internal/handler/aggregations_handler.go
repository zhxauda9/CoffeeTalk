package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"hot-coffee/internal/error_handler"
	"hot-coffee/internal/service"
)

// AggregationHandler handles aggregation-related HTTP requests such as sales, popular items, and search functionality.
type AggregationHandler struct {
	orderService       *service.OrderService
	aggregationService service.AggregationService
	logger             *slog.Logger
}

// NewAggregationHandler creates a new AggregationHandler instance with dependencies injected.
func NewAggregationHandler(orderService *service.OrderService, aggregationService service.AggregationService, logger *slog.Logger) *AggregationHandler {
	return &AggregationHandler{orderService: orderService, aggregationService: aggregationService, logger: logger}
}

// TotalSalesHandler handles requests for the total sales.
func (h *AggregationHandler) TotalSalesHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the HTTP method is GET, otherwise return Method Not Allowed
	if r.Method != http.MethodGet {
		error_handler.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Fetch the total sales data
	totalSales, err := h.orderService.GetTotalSales()
	if err != nil {
		h.logger.Error("Error getting data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Error getting data", http.StatusInternalServerError)
		return
	}

	// Return total sales as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(totalSales)

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

// PopularItemsHandler handles requests to retrieve the popular menu items.
func (h *AggregationHandler) PopularItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch the most popular menu items
	popularItems, err := h.aggregationService.GetPopularMenuItems()
	if err != nil {
		h.logger.Error("Error getting popular items", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Error getting popular items", http.StatusInternalServerError)
		return
	}

	// Return popular items as JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(popularItems)

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

// SearchHandler handles search requests based on query parameters (search, filter, price range).
func (h *AggregationHandler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve query parameters from the URL
	searchQuery := r.URL.Query().Get("q")
	filter := r.URL.Query().Get("filter")
	minPrice := r.URL.Query().Get("minPrice")
	maxPrice := r.URL.Query().Get("maxPrice")

	// Parse minPrice and handle errors
	var MinPrice int
	var err error
	if minPrice != "" {
		MinPrice, err = strconv.Atoi(minPrice)
		if err != nil {
			h.logger.Error("Min Price should be number", "method", r.Method, "url", r.URL)
			error_handler.Error(w, "Min Price should be number", http.StatusBadRequest)
			return
		}
		if MinPrice < 0 {
			h.logger.Error("Min Price should be positive", "method", r.Method, "url", r.URL)
			error_handler.Error(w, service.ErrPriceNotPositive.Error(), http.StatusBadRequest)
			return
		}
	} else {
		MinPrice = -1 // Default value if minPrice is not specified
	}

	// Parse maxPrice and handle errors
	var MaxPrice int
	if maxPrice != "" {
		MaxPrice, err = strconv.Atoi(maxPrice)
		if err != nil {
			h.logger.Error("Max Price should be number", "method", r.Method, "url", r.URL)
			error_handler.Error(w, "Max Price should be number", http.StatusBadRequest)
			return
		}

		if MaxPrice < 0 {
			h.logger.Error("Max Price should be positive", "method", r.Method, "url", r.URL)
			error_handler.Error(w, service.ErrPriceNotPositive.Error(), http.StatusBadRequest)
			return
		}
	} else {
		MaxPrice = -1 // Default value if maxPrice is not specified
	}

	// Perform the search with the given parameters
	searchResult, err := h.aggregationService.Search(searchQuery, MinPrice, MaxPrice, filter)
	if err != nil {
		h.logger.Error("Error searching", "method", r.Method, "url", r.URL, "err", err.Error())
		if err == service.ErrSearchRequired || err == service.ErrWrongFilterOptions || err == service.ErrPriceNotPositive {
			error_handler.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			error_handler.Error(w, fmt.Sprintf("Error searching %v string", searchQuery), http.StatusInternalServerError)
			return
		}
		return
	}

	// Return search results as JSON response
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(searchResult); err != nil {
		h.logger.Error("Could not encode json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not encode request json data", http.StatusInternalServerError)
	}
}

// OrderByPeriod handles requests to retrieve orders by a specific time period (e.g., month/year).
func (h *AggregationHandler) OrderByPeriod(w http.ResponseWriter, r *http.Request) {
	// Retrieve period, month, and year query parameters
	period := r.URL.Query().Get("period")
	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")

	// Fetch orders by the specified time period
	orders, err := h.orderService.GetOrderedItemsByPeriod(period, month, year)
	if err != nil {
		h.logger.Error(err.Error(), "msg", "Error getting orders by time period", "url", r.URL)
		error_handler.Error(w, fmt.Sprintf("Error getting orders by time period. %v", err), http.StatusBadRequest)
		return
	}

	// Return orders as JSON response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		h.logger.Error(err.Error(), "msg", "Failed to encode orders", "url", r.URL)
		error_handler.Error(w, "Failed to encode orders", http.StatusInternalServerError)
		return
	}
}
