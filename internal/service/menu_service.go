package service

import (
	"errors"
	"os"
	"strings"

	"hot-coffee/internal/dal"
	"hot-coffee/models"
)

// MenuService handles operations related to menu items and inventory management.
type MenuService struct {
	menuRepo      dal.MenuRepository      // Repository for interacting with menu items.
	inventoryRepo dal.InventoryRepository // Repository for interacting with inventory items.
}

// NewMenuService creates a new instance of MenuService with the given repositories.
func NewMenuService(menuRepo dal.MenuRepository, inventoryRepo dal.InventoryRepository) *MenuService {
	return &MenuService{menuRepo: menuRepo, inventoryRepo: inventoryRepo}
}

// DeleteMenuItem deletes a menu item by its ID. It also removes the associated image if it exists.
func (s *MenuService) DeleteMenuItem(MenuItemID int) error {
	// First, retrieve the menu item by its ID to check if it exists
	menuItem, err := s.GetMenuItem(MenuItemID)
	if err != nil {
		return err // Return error if item is not found
	}

	// Remove the image associated with the menu item if it exists
	if err := RemoveImage(menuItem.Image); err != nil {
		return err // Return error if image removal fails
	}

	// Finally, delete the menu item from the repository
	return s.menuRepo.DeleteMenuItemRepo(MenuItemID)
}

// UpdateMenuItem updates an existing menu item in the repository.
func (s *MenuService) UpdateMenuItem(menuItem models.MenuItem) error {
	// Call the repository's method to update the menu item
	return s.menuRepo.UpdateMenuItemRepo(menuItem)
}

// MenuCheckByID checks if a menu item exists before adding or deleting it.
func (s *MenuService) MenuCheckByID(MenuItemID int, isDelete bool) error {
	if isDelete {
		// If deleting, check if the menu item exists
		if s.menuRepo.MenuCheckByIDRepo(MenuItemID) {
			return nil // Item exists, so proceed with deletion
		}
		return errors.New("the requested menu item does not exist in menu") // Item not found, return error
	}

	// If adding, check if the menu item already exists
	if s.menuRepo.MenuCheckByIDRepo(MenuItemID) {
		return errors.New("the requested menu item to add already exists in menu") // Item already exists
	}
	return nil // Item does not exist, proceed with addition
}

// IngredientsCheckByID checks if there are enough ingredients for a menu item by its ID.
func (s *MenuService) IngredientsCheckByID(menuItemID int, quantity int) error {
	// Retrieve all menu items
	menuItems, _ := s.menuRepo.GetAll()
	ingredientsNeeded := make(map[int]float64)
	flag := false

	// Iterate over all menu items to find the required ingredients for the given item
	for _, item := range menuItems {
		if item.ID == menuItemID {
			for _, ingr := range item.Ingredients {
				ingredientsNeeded[ingr.IngredientID] += float64(ingr.Quantity) * float64(quantity)
			}
		}
	}

	// Retrieve all inventory items
	inventoryItems, _ := s.inventoryRepo.GetAll()

	// Check if there are sufficient quantities of the ingredients in inventory
	for _, inventoryItem := range inventoryItems {
		if value, exists := ingredientsNeeded[inventoryItem.IngredientID]; exists {
			flag = true
			if value > inventoryItem.Quantity {
				return errors.New("not enough ingredients for item") // Not enough inventory for the item
			}
		}
	}

	// Return error if there are no matching ingredients in the inventory
	if flag {
		return nil
	}
	return errors.New("no ingredients for item in inventory")
}

// IngredientsCheckForNewItem checks if there are enough ingredients in the inventory to add a new menu item.
func (s *MenuService) IngredientsCheckForNewItem(menuItem models.MenuItem) error {
	// Retrieve all inventory items
	inventoryItems, _ := s.inventoryRepo.GetAll()
	count := 0

	// Iterate over the inventory items to check if all ingredients required for the new item are available
	for _, inventoryItem := range inventoryItems {
		for _, ingredients := range menuItem.Ingredients {
			if ingredients.IngredientID == inventoryItem.IngredientID {
				count++
				if ingredients.Quantity > inventoryItem.Quantity {
					return errors.New("not enough ingredients for item") // Not enough inventory for the new item
				}
			}
		}
	}

	// If not all ingredients are found in the inventory, return an error
	if count != len(menuItem.Ingredients) {
		return errors.New("no ingredients for item in inventory")
	}
	return nil
}

// SubtractIngredientsByID subtracts the required ingredients from the inventory when an order is placed.
func (s *MenuService) SubtractIngredientsByID(OrderID int, quantity int) error {
	// First, check if there are enough ingredients for the given order
	if err := s.IngredientsCheckByID(OrderID, quantity); err != nil {
		return errors.New("not enough ingredients or needed ingredients do not exist") // Return error if check fails
	}

	// Calculate the ingredients needed for the order
	ingredients := make(map[int]float64)
	menuItems, _ := s.menuRepo.GetAll()

	for _, item := range menuItems {
		if item.ID == OrderID {
			for _, ingr := range item.Ingredients {
				ingredients[ingr.IngredientID] += float64(ingr.Quantity) * float64(quantity)
			}
		}
	}

	// Subtract the ingredients from the inventory
	return s.inventoryRepo.SubtractIngredients(ingredients)
}

// AddMenuItem adds a new menu item to the repository.
func (s *MenuService) AddMenuItem(menuItem models.MenuItem) error {
	// Call the repository's method to add the menu item
	return s.menuRepo.AddMenuItemRepo(menuItem)
}

// GetMenuItem retrieves a menu item by its ID from the repository.
func (s *MenuService) GetMenuItem(MenuItemID int) (models.MenuItem, error) {
	// Retrieve all menu items and search for the item by ID
	MenuItems, err := s.menuRepo.GetAll()
	if err != nil {
		return models.MenuItem{}, err // Return error if failed to retrieve menu items
	}
	for i, MenuItem := range MenuItems {
		if MenuItem.ID == MenuItemID {
			return MenuItems[i], nil // Return the menu item if found
		}
	}
	// Return an error if the item is not found by its ID
	return models.MenuItem{}, errors.New("could not find menu item by the given id")
}

// GetMenuItems retrieves all menu items from the repository.
func (s *MenuService) GetMenuItems() ([]models.MenuItem, error) {
	// Retrieve all menu items from the repository
	MenuItems, err := s.menuRepo.GetAll()
	if err != nil {
		return []models.MenuItem{}, err // Return error if failed to retrieve menu items
	}
	return MenuItems, err
}

// CheckNewMenu validates the details of a new menu item before adding it to the menu.
func (s *MenuService) CheckNewMenu(MenuItem models.MenuItem) error {
	// Validate that the menu item's name, description, and price are correctly provided
	if strings.TrimSpace(MenuItem.Name) == "" {
		return errors.New("new menu item's Name is empty")
	}
	if strings.TrimSpace(MenuItem.Description) == "" {
		return errors.New("new menu item's Description is empty")
	}
	if MenuItem.Price < 0 {
		return errors.New("new menu item's Price is awkward") // Price should not be negative
	}
	// Validate that each ingredient's quantity is valid (not negative)
	for _, ingredient := range MenuItem.Ingredients {
		if ingredient.Quantity < 0 {
			return errors.New("new menu item's quantity is awkward") // Quantity should not be negative
		}
	}
	return nil // Return nil if all validations pass
}

// UpdateMenuItemImage updates the image of a menu item.
func (s *MenuService) UpdateMenuItemImage(id int, newImagePath string) error {
	// Retrieve the menu item by its ID
	menuItem, err := s.GetMenuItem(id)
	if err != nil {
		return err // Return error if the item is not found
	}

	// Remove the old image associated with the menu item
	if err := RemoveImage(menuItem.Image); err != nil {
		return err // Return error if image removal fails
	}

	// Update the menu item with the new image path
	return s.menuRepo.UpdateMenuItemImageRepo(id, newImagePath)
}

// RemoveImage deletes an image from the filesystem if it's not the default image.
func RemoveImage(imagePath string) error {
	if imagePath != "uploads/default.jpg" {
		// Attempt to remove the image file from the filesystem
		if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
			return err // Return error if the file cannot be removed
		}
	}
	return nil // Return nil if the image removal succeeds or the file does not exist
}
