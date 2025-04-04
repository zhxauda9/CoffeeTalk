package dal

import (
	"database/sql"
	"fmt"

	"hot-coffee/models"
)

// MenuRepository defines methods for interacting with the menu in the database.
type MenuRepository struct {
	db *sql.DB
}

// NewMenuRepository returns a new instance of MenuRepository with the provided database connection.
func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

// GetAll retrieves all menu items from the database, along with their ingredients.
func (repo *MenuRepository) GetAll() ([]models.MenuItem, error) {
	// Query to get all menu items
	queryMenuItems := `
	select ID, Name, Description, Price, Image from menu_items
	`
	rows, err := repo.db.Query(queryMenuItems)
	if err != nil {
		return []models.MenuItem{}, err // Return empty slice if error occurs
	}

	var MenuItems []models.MenuItem
	// Iterate through each menu item in the result set
	for rows.Next() {
		var MenuItem models.MenuItem
		err := rows.Scan(&MenuItem.ID, &MenuItem.Name, &MenuItem.Description, &MenuItem.Price, &MenuItem.Image)
		if err != nil {
			return []models.MenuItem{}, err
		}

		// Get ingredients for each menu item
		var MenuItemIngredients []models.MenuItemIngredient
		queryMenuItemIngredients := `
			select IngredientID, Quantity from menu_item_ingredients where MenuID = $1
		`
		rows1, err := repo.db.Query(queryMenuItemIngredients, MenuItem.ID)
		if err != nil {
			return []models.MenuItem{}, err
		}
		// Iterate through ingredients for each menu item
		for rows1.Next() {
			var MenuItemIngredient models.MenuItemIngredient
			rows1.Scan(&MenuItemIngredient.IngredientID, &MenuItemIngredient.Quantity)
			MenuItemIngredients = append(MenuItemIngredients, MenuItemIngredient)
		}
		// Assign ingredients to the MenuItem
		MenuItem.Ingredients = MenuItemIngredients
		MenuItems = append(MenuItems, MenuItem)
	}
	return MenuItems, nil // Return all menu items
}

// Exists checks whether a menu item with the given ID exists in the database.
func (repo *MenuRepository) Exists(itemID int) bool {
	items, _ := repo.GetAll() // Get all menu items
	for _, item := range items {
		if item.ID == itemID { // Check if item ID matches
			return true
		}
	}
	return false // Return false if not found
}

// DeleteMenuItemRepo deletes a menu item from the database using the given ID.
func (repo *MenuRepository) DeleteMenuItemRepo(MenuItemID int) error {
	queryDeleteMenuItem := `
	delete from menu_items
	where ID = $1
	`
	// Execute delete query
	_, err := repo.db.Exec(queryDeleteMenuItem, MenuItemID)
	if err != nil {
		return err // Return error if deletion fails
	}
	return nil // Return nil if deletion is successful
}

// UpdateMenuItemRepo updates the details of an existing menu item in the database.
func (repo *MenuRepository) UpdateMenuItemRepo(menuItem models.MenuItem) error {
	// Query to update menu item
	queryUpdateMenu := `
	update menu_items
	set Name = $1, Description = $2, Price = $3, Image=$4
	where ID = $5
	`
	// Execute the update query
	_, err := repo.db.Exec(queryUpdateMenu, menuItem.Name, menuItem.Description, menuItem.Price, menuItem.Image, menuItem.ID)
	if err != nil {
		return err // Return error if update fails
	}

	// Delete existing ingredients for this menu item
	queryUpdateMenuIngredients1 := `
			delete from menu_item_ingredients 
			where MenuID = $1
		`
	_, err = repo.db.Exec(queryUpdateMenuIngredients1, menuItem.ID)
	if err != nil {
		return err // Return error if ingredient deletion fails
	}

	// Insert new ingredients for this menu item
	for _, v := range menuItem.Ingredients {
		queryUpdateMenuIngredients2 := `
			insert into menu_item_ingredients (MenuID, IngredientID, Quantity) values
			($1, $2, $3)
		`
		_, err = repo.db.Exec(queryUpdateMenuIngredients2, menuItem.ID, v.IngredientID, v.Quantity)
		if err != nil {
			return err // Return error if ingredient insertion fails
		}
	}
	return nil // Return nil if update is successful
}

// UpdateMenuItemImageRepo updates the image path of a menu item using the provided ID and image path.
func (r *MenuRepository) UpdateMenuItemImageRepo(id int, imagePath string) error {
	query := "update menu_items set Image = $1 where ID = $2"

	// Execute query to update image
	_, err := r.db.Exec(query, imagePath, id)
	if err != nil {
		return fmt.Errorf("could not update image: %w", err) // Return error if image update fails
	}

	return nil // Return nil if image update is successful
}

// AddMenuItemRepo adds a new menu item to the database.
func (repo *MenuRepository) AddMenuItemRepo(menuItem models.MenuItem) error {
	// Query to insert new menu item
	queryAddItem := `
		INSERT INTO menu_items (Name, Description, Price, Image) 
		VALUES ($1, $2, $3, $4) RETURNING ID
	`
	var newID int
	// Execute the insert query and get the new ID
	err := repo.db.QueryRow(queryAddItem, menuItem.Name, menuItem.Description, menuItem.Price, menuItem.Image).Scan(&newID)
	if err != nil {
		return err // Return error if insertion fails
	}

	menuItem.ID = newID // Set the ID of the new menu item

	// Add ingredients for the new menu item
	for _, v := range menuItem.Ingredients {
		queryAddItemIngredients := `
			INSERT INTO menu_item_ingredients (MenuID, IngredientID, Quantity) 
			VALUES ($1, $2, $3)
		`
		_, err = repo.db.Exec(queryAddItemIngredients, menuItem.ID, v.IngredientID, v.Quantity)
		if err != nil {
			return err // Return error if ingredient insertion fails
		}
	}
	return nil // Return nil if item and ingredients are added successfully
}

// MenuCheckByIDRepo checks if a menu item exists by its ID.
func (repo *MenuRepository) MenuCheckByIDRepo(ID int) bool {
	queryIfExists := `
	select ID from menu_items where ID = $1
	`
	rows, err := repo.db.Query(queryIfExists, ID)
	if err != nil {
		return false // Return false if query fails
	}
	return rows.Next() // Return true if the menu item exists
}
