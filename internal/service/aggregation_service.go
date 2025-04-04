package service

import (
	"errors"
	"strings"

	"hot-coffee/internal/dal"
	"hot-coffee/models"
)

// Custom errors for specific cases.
var (
	ErrWrongFilterOptions = errors.New("no such filter. Available filters: orders, menu, all")
	ErrSearchRequired     = errors.New("search query string is required")
	ErrPriceNotPositive   = errors.New("minPrice and maxPrice must be positive")
)

// AggregationService defines the interface for aggregation-related operations.
type AggregationService interface {
	// GetPopularMenuItems retrieves popular menu items.
	GetPopularMenuItems() (models.PopularItems, error)
	// Search allows searching menu items, orders, or both with filters.
	Search(searchQuery string, minPrice, maxPrice int, filter string) (models.SearchResult, error)
}

// AggregationServiceImpl implements the AggregationService interface.
type AggregationServiceImpl struct {
	searchRepo dal.ReportRespository // Repository for accessing report-related data.
}

// NewAggregationService creates and returns a new instance of AggregationServiceImpl.
func NewAggregationService(searchRepo dal.ReportRespository) *AggregationServiceImpl {
	return &AggregationServiceImpl{searchRepo: searchRepo}
}

// GetPopularMenuItems retrieves the most popular menu items.
func (s *AggregationServiceImpl) GetPopularMenuItems() (models.PopularItems, error) {
	// Fetch popular menu items from the repository.
	popItms, err := s.searchRepo.GetPopularMenuItems()
	// Return a result struct with the popular items.
	res := models.PopularItems{
		Items: popItms,
	}
	return res, err
}

// Search performs a search for menu items and orders based on query and filter parameters.
func (s *AggregationServiceImpl) Search(searchQuery string, minPrice, maxPrice int, filter string) (models.SearchResult, error) {
	// Check if the search query is empty.
	if searchQuery == "" {
		return models.SearchResult{}, ErrSearchRequired // Return error if search query is missing.
	}
	// Validate that minPrice and maxPrice are positive values or -1.
	if minPrice < -1 || maxPrice < -1 {
		return models.SearchResult{}, ErrPriceNotPositive // Return error if price range is invalid.
	}

	// Validate the search filter options.
	isOrders, isMenu, err := validateSearchFilters(filter)
	if err != nil {
		return models.SearchResult{}, err // Return error if filter validation fails.
	}

	// Initialize slices for menu items and orders.
	var menuItems []models.SearchMenuItem
	if isMenu {
		// Fetch menu items based on the search query and price range.
		menuItems, err = s.searchRepo.SearchMenuItems(searchQuery, minPrice, maxPrice)
		if err != nil {
			return models.SearchResult{}, err // Return error if fetching menu items fails.
		}
	}

	var orders []models.SearchOrderResult
	if isOrders {
		// Fetch orders based on the search query.
		orders, err = s.searchRepo.SearchOrders(searchQuery)
		if err != nil {
			return models.SearchResult{}, err // Return error if fetching orders fails.
		}
	}

	// Compile the search results into a SearchResult struct.
	result := models.SearchResult{
		MenuItems:    menuItems,
		Orders:       orders,
		TotalMatches: len(menuItems) + len(orders), // Total matches are the sum of menu items and orders.
	}
	return result, nil // Return the search results.
}

// validateSearchFilters validates the filter options passed in the search.
func validateSearchFilters(filter string) (bool, bool, error) {
	var args []string

	// If no filter is provided, assume we are searching both menu and orders.
	if filter == "" {
		return true, true, nil
	}

	var isOrders, isMenu bool

	// Split the filter string by commas.
	args = strings.Split(filter, ",")
	for _, v := range args {
		// Ensure the filter options are valid (only "orders", "menu", or "all").
		if v != "orders" && v != "menu" && v != "all" {
			return false, false, ErrWrongFilterOptions // Return error if invalid filter is found.
		}
		if v == "orders" {
			isOrders = true // Mark that orders should be included in the search.
		} else if v == "menu" {
			isMenu = true // Mark that menu items should be included in the search.
		} else if v == "all" {
			// If "all" is specified, include both menu items and orders in the search.
			isOrders = true
			isMenu = true
		}
	}
	return isOrders, isMenu, nil // Return the selected filter options.
}
