package service

import (
	"errors"
	"strconv"

	"hot-coffee/internal/dal"
	"hot-coffee/models"
)

// InventoryService manages the operations related to inventory items.
type InventoryService struct {
	inventoryRepo dal.InventoryRepository // Repository for interacting with inventory items.
}

// NewInventoryService creates and returns a new instance of InventoryService with the given repository.
func NewInventoryService(inventoryRepo dal.InventoryRepository) *InventoryService {
	return &InventoryService{inventoryRepo: inventoryRepo}
}

// AddInventoryItem adds a new inventory item to the repository.
func (s *InventoryService) AddInventoryItem(item models.InventoryItem) error {
	// Add the inventory item using the repository's method
	return s.inventoryRepo.AddInventoryItemRepo(item)
}

// GetAllInventoryItems retrieves all inventory items from the repository.
func (s *InventoryService) GetAllInventoryItems() ([]models.InventoryItem, error) {
	// Fetch all inventory items from the repository
	items, err := s.inventoryRepo.GetAll()
	if err != nil {
		return []models.InventoryItem{}, nil // Return an empty list if an error occurs
	}
	return items, nil // Return the inventory items
}

// GetItem retrieves a specific inventory item by its ID.
func (s *InventoryService) GetItem(id int) (models.InventoryItem, error) {
	// Fetch all inventory items from the repository
	items, err := s.inventoryRepo.GetAll()
	if err != nil {
		return models.InventoryItem{}, err // Return error if fetching inventory items fails
	}

	// Search for the item by its IngredientID
	for _, item := range items {
		if item.IngredientID == id {
			return item, nil // Return the item if found
		}
	}
	// Return an error if the inventory item does not exist
	return models.InventoryItem{}, errors.New("inventory item does not exist")
}

// UpdateItem updates an existing inventory item identified by its ID.
func (s *InventoryService) UpdateItem(id int, newItem models.InventoryItem) error {
	// Check if the inventory item exists before updating it
	if !s.inventoryRepo.Exists(id) {
		return errors.New("inventory item does not exist") // Return error if the item doesn't exist
	}
	// Update the inventory item in the repository
	return s.inventoryRepo.UpdateItemRepo(id, newItem)
}

// DeleteItem deletes an inventory item by its ID.
func (s *InventoryService) DeleteItem(id int) error {
	// Check if the inventory item exists before deleting it
	if !s.inventoryRepo.Exists(id) {
		return errors.New("inventory item does not exist") // Return error if the item doesn't exist
	}
	// Delete the inventory item from the repository
	return s.inventoryRepo.DeleteItemRepo(id)
}

// Exists checks if an inventory item exists by its ID.
func (s *InventoryService) Exists(id int) bool {
	// Return whether the item exists in the repository
	return s.inventoryRepo.Exists(id)
}

// GetLeftOvers retrieves the inventory items with their remaining quantities, with sorting and pagination options.
func (s *InventoryService) GetLeftOvers(sortBy, page, pageSize string) (map[string]any, error) {
	// Set default sorting parameter to "price" if not provided
	if sortBy == "" {
		sortBy = "price"
	}

	// Set default page number to 1 if not provided
	if page == "" {
		page = "1"
	}
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 0 {
		return nil, errors.New("invalid page parameter. page must be a positive integer") // Return error if page is invalid
	}

	// Set default page size to 10 if not provided
	if pageSize == "" {
		pageSize = "10"
	}
	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 0 {
		return nil, errors.New("invalid pageSize parameter. pageSize must be a positive integer") // Return error if pageSize is invalid
	}

	// Validate the sortBy parameter, ensuring it's either "price" or "quantity"
	if sortBy != "price" && sortBy != "quantity" {
		return nil, errors.New("invalid sortBy value, must be 'price' or 'quantity'") // Return error if sortBy is invalid
	}

	// Retrieve the inventory leftovers based on the sorting and pagination parameters
	return s.inventoryRepo.GetLeftOvers(sortBy, page, pageSize)
}
