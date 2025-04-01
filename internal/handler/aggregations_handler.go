package handler

import (
	"encoding/json"
	"fmt"
	"hot-coffee/internal/error_handler"
	"log/slog"
	"net/http"
	"strconv"

	"hot-coffee/internal/service"
)

type AggregationHandler struct {
	orderService       *service.OrderService
	aggregationService service.AggregationService
	logger             *slog.Logger
}

func NewAggregationHandler(orderService *service.OrderService, aggregationService service.AggregationService, logger *slog.Logger) *AggregationHandler {
	return &AggregationHandler{orderService: orderService, aggregationService: aggregationService, logger: logger}
}

func (h *AggregationHandler) TotalSalesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		error_handler.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	totalSales, err := h.orderService.GetTotalSales()
	if err != nil {
		h.logger.Error("Error getting data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Error getting data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(totalSales)

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *AggregationHandler) PopularItemsHandler(w http.ResponseWriter, r *http.Request) {
	popularItems, err := h.aggregationService.GetPopularMenuItems()
	if err != nil {
		h.logger.Error("Error getting popular items", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Error getting popular items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(popularItems)

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *AggregationHandler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("q")
	filter := r.URL.Query().Get("filter")
	minPrice := r.URL.Query().Get("minPrice")
	maxPrice := r.URL.Query().Get("maxPrice")

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
		MinPrice = -1
	}

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
		MaxPrice = -1
	}

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

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(searchResult); err != nil {
		h.logger.Error("Could not encode json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not encode request json data", http.StatusInternalServerError)
	}
}

func (h *AggregationHandler) OrderByPeriod(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")

	orders, err := h.orderService.GetOrderedItemsByPeriod(period, month, year)
	if err != nil {
		h.logger.Error(err.Error(), "msg", "Error getting orders by time period", "url", r.URL)
		error_handler.Error(w, fmt.Sprintf("Error getting orders by time period. %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		h.logger.Error(err.Error(), "msg", "Failed to encode orders", "url", r.URL)
		error_handler.Error(w, "Failed to encode orders", http.StatusInternalServerError)
		return
	}
}
