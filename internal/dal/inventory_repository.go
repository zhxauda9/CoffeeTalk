package dal

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"hot-coffee/models"
)

// InventoryRepository is responsible for interacting with the inventory data in the database.
type InventoryRepository struct {
	db *sql.DB
}

// NewInventoryRepository creates and returns a new instance of InventoryRepository.
func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// GetAll retrieves all inventory items from the database.
func (repo *InventoryRepository) GetAll() ([]models.InventoryItem, error) {
	// SQL query to get all inventory items
	queryGetIngridients := `
	select IngredientID, Name, Quantity, Unit from inventory
	`
	rows, err := repo.db.Query(queryGetIngridients)
	if err != nil {
		return []models.InventoryItem{}, err // Return empty slice in case of error
	}

	var InventoryItems []models.InventoryItem
	// Iterate through all rows returned by the query
	for rows.Next() {
		var InventoryItem models.InventoryItem
		err = rows.Scan(&InventoryItem.IngredientID, &InventoryItem.Name, &InventoryItem.Quantity, &InventoryItem.Unit)
		if err != nil {
			return []models.InventoryItem{}, nil // Return nil if scanning fails
		}
		// Append each InventoryItem to the InventoryItems slice
		InventoryItems = append(InventoryItems, InventoryItem)
	}
	return InventoryItems, nil // Return the list of all inventory items
}

// Exists checks if an inventory item with the given ID exists in the database.
func (repo *InventoryRepository) Exists(ID int) bool {
	// SQL query to check if an inventory item with the provided ID exists
	queryIfExists := `
	select IngredientID from inventory where IngredientID = $1
	`
	rows, err := repo.db.Query(queryIfExists, ID)
	if err != nil {
		return false // Return false in case of query error
	}
	return rows.Next() // Return true if the row exists, else false
}

// SubtractIngredients subtracts quantities of ingredients from the inventory based on the provided map.
func (repo *InventoryRepository) SubtractIngredients(ingredients map[int]float64) error {
	// Iterate over all ingredients in the map (ingredient ID -> quantity to subtract)
	for key, value := range ingredients {
		// SQL query to update the inventory by subtracting the specified quantity
		queryToSubtract := `
	        update inventory
	        set Quantity  = Quantity - $1
	        where IngredientID = $2
	    `
		_, err := repo.db.Exec(queryToSubtract, value, key)
		if err != nil {
			return err // Return error if any query fails
		}
	}
	return nil // Return nil if all quantities are successfully subtracted
}

// AddInventoryItemRepo adds a new inventory item to the database.
func (repo *InventoryRepository) AddInventoryItemRepo(item models.InventoryItem) error {
	// SQL query to insert a new inventory item into the database
	queryToAddInventory := `
	insert into inventory (Name, Quantity, Unit) values
	($1, $2, $3)
	`
	_, err := repo.db.Exec(queryToAddInventory, item.Name, item.Quantity, item.Unit)
	if err != nil {
		return err // Return error if insertion fails
	}
	return nil // Return nil if item is successfully added
}

// UpdateItemRepo updates an existing inventory item's details in the database.
func (repo *InventoryRepository) UpdateItemRepo(id int, newItem models.InventoryItem) error {
	// SQL query to update an inventory item based on the provided ID
	queryToUpdate := `
	update inventory
	set Quantity = $1, Name = $2, Unit = $3
	where IngredientID = $4
	`
	_, err := repo.db.Exec(queryToUpdate, newItem.Quantity, newItem.Name, newItem.Unit, id)
	if err != nil {
		return err // Return error if update fails
	}
	return nil // Return nil if update is successful
}

// DeleteItemRepo deletes an inventory item based on its ID.
func (repo *InventoryRepository) DeleteItemRepo(id int) error {
	// SQL query to delete an inventory item using the given ID
	queryToDelete := `
	delete from inventory
	where IngredientID = $1
	`
	_, err := repo.db.Exec(queryToDelete, id)
	if err != nil {
		fmt.Println("Delete error:", err)
		return err // Return error if deletion fails
	}
	return nil // Return nil if item is successfully deleted
}

// GetLeftOvers retrieves a paginated list of inventory items, sorted by a specified field (either 'price' or 'quantity').
func (repo *InventoryRepository) GetLeftOvers(sortBy, page, pageSize string) (map[string]any, error) {
	// Convert page and pageSize to integers
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum <= 0 {
		pageNum = 1 // Default to page 1 if conversion fails or invalid value
	}
	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum <= 0 {
		pageSizeNum = 10 // Default to 10 items per page if conversion fails or invalid value
	}
	// Calculate the offset for pagination
	offset := (pageNum - 1) * pageSizeNum

	// Base query to retrieve inventory items
	query := `
        SELECT i.IngredientID, i.Name, i.Quantity, i.Unit
        FROM inventory i
    `

	// Sort the query based on the sortBy parameter
	switch sortBy {
	case "price":
		query += " ORDER BY i.Price"
	case "quantity":
		query += " ORDER BY i.Quantity"
	default:
		return nil, errors.New("invalid sortBy value, must be 'price' or 'quantity'") // Return error if invalid sortBy value
	}

	// Add pagination to the query
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSizeNum, offset)

	var leftovers []map[string]any
	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err) // Return error if the query fails
	}
	defer rows.Close()

	// Process the rows returned from the query
	for rows.Next() {
		var ingredientID int
		var name string
		var quantity int
		var unit string
		if err := rows.Scan(&ingredientID, &name, &quantity, &unit); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err) // Return error if row scan fails
		}

		// Append each row data to the leftovers slice
		leftovers = append(leftovers, map[string]any{
			"ingredientID": ingredientID,
			"name":         name,
			"quantity":     quantity,
			"unit":         unit,
		})
	}

	// Check for any iteration error
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	// Get the total number of items in the inventory
	var totalItems int
	err = repo.db.QueryRow("SELECT COUNT(*) FROM inventory").Scan(&totalItems)
	if err != nil {
		return nil, fmt.Errorf("failed to count total items: %v", err)
	}

	// Calculate the total number of pages and check if there's a next page
	totalPages := (totalItems + pageSizeNum - 1) / pageSizeNum
	hasNextPage := pageNum < totalPages

	// Prepare the response object with pagination info
	response := map[string]any{
		"currentPage": pageNum,
		"hasNextPage": hasNextPage,
		"pageSize":    pageSizeNum,
		"totalPages":  totalPages,
		"data":        leftovers,
	}

	return response, nil // Return the paginated inventory data
}
