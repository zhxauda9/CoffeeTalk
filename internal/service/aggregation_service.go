package service

import (
	"errors"
	"strings"

	"hot-coffee/internal/dal"
	"hot-coffee/models"
)

var (
	ErrWrongFilterOptions = errors.New("no such filter. Available filters: orders, menu, all")
	ErrSearchRequired     = errors.New("search query string is required")
	ErrPriceNotPositive   = errors.New("minPrice and maxPrice must be postive")
)

type AggregationService interface {
	GetPopularMenuItems() (models.PopularItems, error)
	Search(searchQuery string, minPrice, maxPrice int, filter string) (models.SearchResult, error)
}

type AggregationServiceImpl struct {
	searchRepo dal.ReportRespository
}

func NewAggregationService(searchRepo dal.ReportRespository) *AggregationServiceImpl {
	return &AggregationServiceImpl{searchRepo: searchRepo}
}

func (s *AggregationServiceImpl) GetPopularMenuItems() (models.PopularItems, error) {
	popItms, err := s.searchRepo.GetPopularMenuItems()
	res := models.PopularItems{
		Items: popItms,
	}
	return res, err
}

func (s *AggregationServiceImpl) Search(searchQuery string, minPrice, maxPrice int, filter string) (models.SearchResult, error) {
	var err error

	if searchQuery == "" {
		return models.SearchResult{}, ErrSearchRequired
	}
	if minPrice < -1 || maxPrice < -1 {
		return models.SearchResult{}, ErrPriceNotPositive
	}

	isOrders, isMenu, err := validateSearchFilters(filter)
	if err != nil {
		return models.SearchResult{}, err
	}

	var menuItems []models.SearchMenuItem
	if isMenu {
		menuItems, err = s.searchRepo.SearchMenuItems(searchQuery, minPrice, maxPrice)
		if err != nil {
			return models.SearchResult{}, err
		}
	}

	var orders []models.SearchOrderResult
	if isOrders {
		orders, err = s.searchRepo.SearchOrders(searchQuery)
		if err != nil {
			return models.SearchResult{}, err
		}
	}

	result := models.SearchResult{
		MenuItems:    menuItems,
		Orders:       orders,
		TotalMatches: len(menuItems) + len(orders),
	}
	return result, nil
}

func validateSearchFilters(filter string) (bool, bool, error) {
	var args []string

	if filter == "" {
		return true, true, nil
	}

	var isOrders, isMenu bool

	args = strings.Split(filter, ",")
	for _, v := range args {
		if v != "orders" && v != "menu" && v != "all" {
			return false, false, ErrWrongFilterOptions
		}
		if v == "orders" {
			isOrders = true
		} else if v == "menu" {
			isMenu = true
		} else if v == "all" {
			isOrders = true
			isMenu = true
		}
	}
	return isOrders, isMenu, nil
}
