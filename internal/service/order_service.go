package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"hot-coffee/internal/dal"
	"hot-coffee/models"
)

// OrderService struct contains repositories for orders, menu, and inventory.
type OrderService struct {
	orderRepo     dal.OrderRepository
	menuRepo      dal.MenuRepository
	inventoryRepo dal.InventoryRepository
}

// NewOrderService is a constructor function to create a new instance of OrderService.
func NewOrderService(orderRepo dal.OrderRepository, menuRepo dal.MenuRepository, inventoryRepo dal.InventoryRepository) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		menuRepo:      menuRepo,
		inventoryRepo: inventoryRepo,
	}
}

// AddOrder processes a single order by validating and adding it to the repository.
func (s *OrderService) AddOrder(order models.Order) (models.BatchOrderInfo, []models.BatchOrderInventoryUpdate, error) {
	// Validate the order to ensure the provided data is correct
	if err := validateOrder(order); err != nil {
		// If validation fails, return the error message and order rejection status
		return models.BatchOrderInfo{
			OrderID:      order.ID,
			CustomerName: order.CustomerName,
			Status:       models.StatusOrderRejected,
			Reason:       err.Error(),
			Total:        0,
		}, []models.BatchOrderInventoryUpdate{}, err
	}

	// If validation passes, proceed to add the order to the repository
	return s.orderRepo.Add(order)
}

// BulkOrders processes multiple orders in a batch, updating the inventory and sales summary.
func (s *OrderService) BulkOrders(orders []models.Order) (models.BatchOrdersResponce, error) {
	proccesedOrdersInfo := []models.BatchOrderInfo{} // Store info about each processed order
	summary := models.BatchOrderSummary{
		TotalOrders: len(orders),
	}

	// Map to track inventory updates across multiple orders
	invCheckMap := make(map[int]models.BatchOrderInventoryUpdate)
	for _, order := range orders {
		// Process each order individually
		orderInfo, inventoryInfo, err := s.AddOrder(order)
		if err != nil {
			log.Printf("Error: %v", err)
		}

		// Update summary based on the order status
		if orderInfo.Status == models.StatusOrderAccepted {
			summary.Accepted++
		} else {
			summary.Rejected++
		}
		summary.TotalRevenue += orderInfo.Total
		proccesedOrdersInfo = append(proccesedOrdersInfo, orderInfo)

		// Update the inventory tracking map
		for _, v := range inventoryInfo {
			if value, ok := invCheckMap[v.IngredientID]; ok {
				// If ingredient already exists in the map, accumulate the quantity used
				v.Quantity_used += value.Quantity_used
				invCheckMap[v.IngredientID] = v
			} else {
				invCheckMap[v.IngredientID] = v
			}
		}

		// Close the order in the repository once it has been processed
		err = s.orderRepo.CloseOrderRepo(orderInfo.OrderID)
		if err != nil && err != models.ErrOrderNotFound {
			return models.BatchOrdersResponce{}, err
		}
	}

	// Append the inventory updates to the summary
	for _, val := range invCheckMap {
		summary.InventoryUpdates = append(summary.InventoryUpdates, val)
	}

	// Return the processed orders and summary
	result := models.BatchOrdersResponce{
		Processed_orders: proccesedOrdersInfo,
		Summary:          summary,
	}
	return result, nil
}

// GetAllOrders retrieves all orders from the order repository.
func (s *OrderService) GetAllOrders() ([]models.Order, error) {
	return s.orderRepo.GetAll()
}

// GetOrder retrieves a specific order by its ID from the repository.
func (s *OrderService) GetOrder(OrderID int) (models.Order, error) {
	return s.orderRepo.GetOrderByID(OrderID)
}

// UpdateOrder updates an existing order in the repository.
func (s *OrderService) UpdateOrder(updatedOrder models.Order, OrderID string) error {
	// Validate the updated order
	if err := validateOrder(updatedOrder); err != nil {
		return err
	}
	// Save the updated order to the repository
	return s.orderRepo.SaveUpdatedOrder(updatedOrder, OrderID)
}

// GetTotalSales calculates the total sales by summing up the quantities of all items in all orders.
func (s *OrderService) GetTotalSales() (models.TotalSales, error) {
	existingOrders, err := s.orderRepo.GetAll()
	if err != nil {
		return models.TotalSales{}, err
	}

	totalSales := models.TotalSales{}

	// Sum the quantities of items in each order
	for _, order := range existingOrders {
		for _, item := range order.Items {
			totalSales.TotalSales += item.Quantity
		}
	}
	return totalSales, nil
}

// DeleteOrderByID deletes a specific order by its ID.
func (s *OrderService) DeleteOrderByID(OrderID int) error {
	return s.orderRepo.DeleteOrder(OrderID)
}

// CloseOrder marks an order as closed in the repository.
func (s *OrderService) CloseOrder(OrderID int) error {
	return s.orderRepo.CloseOrderRepo(OrderID)
}

// GetNumberOfItems returns the number of ordered items between the provided date range.
func (s *OrderService) GetNumberOfItems(startDate, endDate string) (map[string]int, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid time format of startDate")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid time format of endDate")
	}

	return s.orderRepo.GetNumberOfItems(start, end)
}

// GetOrderedItemsByPeriod retrieves ordered items within a specific period (day or month).
func (s *OrderService) GetOrderedItemsByPeriod(period, month, year string) (map[string]interface{}, error) {
	if period == "" {
		return nil, fmt.Errorf("period is required")
	}

	// Validate period and month/year parameters
	if period == "day" {
		if month == "" {
			return nil, fmt.Errorf("period equal to 'day', but month not provided")
		}
	} else if period == "month" {
		if year == "" {
			return nil, fmt.Errorf("period equal to 'month', but year not provided")
		}
	} else {
		return nil, fmt.Errorf("invalid period value, must be 'day' or 'month'")
	}

	// Handle "day" period logic
	if period == "day" {
		monthInt := getMonthNumber(strings.ToLower(month))
		if monthInt == -1 {
			return nil, fmt.Errorf("%s, month does not exist", month)
		}

		var yearInt int
		if year == "" {
			yearInt = -1 // If year is not provided, set to -1 (i.e., no specific year filter)
		} else {
			n, err := strconv.Atoi(year)
			if err != nil {
				return nil, fmt.Errorf("year must be a number")
			}
			yearInt = n
		}

		return s.orderRepo.OrderedItemsByDay(monthInt, yearInt)
	} else if period == "month" {
		// Handle "month" period logic
		yearInt, err := strconv.Atoi(year)
		if err != nil {
			return nil, fmt.Errorf("year should be number")
		}
		return s.orderRepo.OrderedItemsByMonth(yearInt)
	}

	// Return an error if invalid period inputs
	return nil, fmt.Errorf("invalid inputs. Period: %v, Month: %s, Year: %s", period, month, year)
}

// getMonthNumber maps month names to month numbers.
func getMonthNumber(month string) int {
	months := map[string]int{
		"january":   1,
		"february":  2,
		"march":     3,
		"april":     4,
		"may":       5,
		"june":      6,
		"july":      7,
		"august":    8,
		"september": 9,
		"october":   10,
		"november":  11,
		"december":  12,
	}

	v, ok := months[strings.ToLower(month)]
	if !ok {
		return -1 // Return -1 if month is invalid
	}
	return v
}

// validateOrder ensures that an order is valid before it is processed.
func validateOrder(order models.Order) error {
	if order.Items == nil {
		return errors.New("no items provided. Array of items it required")
	}

	if strings.TrimSpace(order.CustomerName) == "" {
		return errors.New("customer name is required")
	}

	// Ensure that each item has a valid quantity
	for _, order := range order.Items {
		if order.Quantity < 1 {
			return errors.New("quantity a product must be greater than zero")
		}
	}

	return nil
}
