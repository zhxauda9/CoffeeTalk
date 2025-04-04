package dal

import (
	"database/sql"
	"fmt"
	"math"

	"hot-coffee/models"

	"github.com/lib/pq"
)

// ReportRespository is the interface defining methods for fetching reports like popular menu items and search results for orders and menu items.
type ReportRespository interface {
	GetPopularMenuItems() ([]models.PopularItem, error)
	SearchOrders(searchQuery string) ([]models.SearchOrderResult, error)
	SearchMenuItems(searchQuery string, minPrice, maxPrice int) ([]models.SearchMenuItem, error)
}

// ReportRespositoryImpl is the concrete implementation of the ReportRespository interface.
type ReportRespositoryImpl struct {
	db *sql.DB // The database connection
}

// NewReportRespository creates a new instance of ReportRespositoryImpl, initialized with a database connection.
func NewReportRespository(db *sql.DB) *ReportRespositoryImpl {
	return &ReportRespositoryImpl{db: db}
}

// GetPopularMenuItems retrieves the most popular menu items based on the total quantity sold.
func (repo *ReportRespositoryImpl) GetPopularMenuItems() ([]models.PopularItem, error) {
	// SQL query to get the most popular menu items based on total quantity sold
	query := `
        SELECT oi.productid, mi.name, mi.description, SUM(quantity) as total, mi.image
        FROM order_items oi
        JOIN menu_items mi on oi.productid = mi.ID
        GROUP BY oi.productid, mi.name, mi.description, mi.image
        ORDER BY total DESC
    `
	// Execute the query
	rows, err := repo.db.Query(query)
	if err != nil {
		return []models.PopularItem{}, fmt.Errorf("error getting popular items %v", err)
	}
	defer rows.Close() // Ensure the rows are closed after processing

	var result []models.PopularItem
	// Loop through the query results and map them to PopularItem structs
	for rows.Next() {
		var item models.PopularItem
		if err := rows.Scan(&item.ProductID, &item.Name, &item.Description, &item.Quantity, &item.Image); err != nil {
			return nil, err // Return error if scanning fails
		}
		result = append(result, item) // Append the item to the result
	}

	return result, nil
}

// SearchOrders performs a full-text search on orders based on the customer name and menu items.
func (repo *ReportRespositoryImpl) SearchOrders(searchQuery string) ([]models.SearchOrderResult, error) {
	// SQL query to search orders based on customer name and menu items, using full-text search for relevance
	query := `
		SELECT 
			ord.ID, 
			ord.CustomerName, 
			ARRAY_AGG(mi.Name) AS items, 
			SUM(mi.Price) AS total,
			ts_rank(
				to_tsvector(ord.CustomerName || ' ' || STRING_AGG(mi.Name, ' ')), 
				websearch_to_tsquery($1)
			) AS relevance
		FROM orders ord
		JOIN order_items oi ON ord.ID = oi.OrderID
		JOIN menu_items mi ON oi.ProductID = mi.ID
		GROUP BY ord.ID, ord.CustomerName
		HAVING to_tsvector(ord.CustomerName || ' ' || STRING_AGG(mi.Name, ' ')) @@ websearch_to_tsquery($1)
		ORDER BY relevance DESC;
	`

	// Execute the query with the search query as a parameter
	rows, err := repo.db.Query(query, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("error searching %v in orders: %v", searchQuery, err)
	}
	defer rows.Close()

	var result []models.SearchOrderResult
	// Loop through the results and scan them into SearchOrderResult structs
	for rows.Next() {
		var item models.SearchOrderResult
		if err := rows.Scan(&item.ID, &item.CustomerName, pq.Array(&item.Items), &item.Total, &item.Relevance); err != nil {
			return nil, err // Return error if scanning fails
		}
		item.Relevance = math.Round(item.Relevance*100) / 100 // Round relevance to two decimal places
		result = append(result, item)                         // Append the result to the slice
	}
	return result, nil
}

// SearchMenuItems performs a full-text search on menu items based on the name and description, and supports filtering by price.
func (repo *ReportRespositoryImpl) SearchMenuItems(searchQuery string, minPrice, maxPrice int) ([]models.SearchMenuItem, error) {
	// SQL query to search menu items based on name and description, using full-text search for relevance
	query := `
		SELECT 
			id, name, description, price,
			ts_rank(to_tsvector(name || ' ' || COALESCE(description, '')), websearch_to_tsquery($1)) as relevance
		FROM menu_items
		WHERE to_tsvector(name || ' ' || COALESCE(description, '')) @@ websearch_to_tsquery($1)
	`
	// Parameters for the query
	args := []interface{}{searchQuery}
	argIndex := 2

	// Add price filtering if the price is specified (not -1)
	if minPrice != -1 {
		query += fmt.Sprintf(" AND price >= $%d", argIndex)
		args = append(args, minPrice)
		argIndex++
	}

	if maxPrice != -1 {
		query += fmt.Sprintf(" AND price <= $%d", argIndex)
		args = append(args, maxPrice)
		argIndex++
	}

	query += " ORDER BY relevance DESC;" // Order by relevance score

	// Execute the query with the parameters
	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error searching %v in menu items: %v", searchQuery, err)
	}
	defer rows.Close()

	var result []models.SearchMenuItem
	// Loop through the results and scan them into SearchMenuItem structs
	for rows.Next() {
		var item models.SearchMenuItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Relevance); err != nil {
			return nil, err // Return error if scanning fails
		}
		item.Relevance = math.Round(item.Relevance*100) / 100 // Round relevance to two decimal places
		result = append(result, item)                         // Append the result to the slice
	}

	if err = rows.Err(); err != nil {
		return nil, err // Return error if there was an issue with the row iteration
	}

	return result, nil
}
