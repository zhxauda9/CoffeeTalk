package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hot-coffee/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (repo *OrderRepository) Add(order models.Order) (models.BatchOrderInfo, []models.BatchOrderInventoryUpdate, error) {
	processInfo := models.BatchOrderInfo{
		CustomerName: order.CustomerName,
		Status:       models.StatusOrderRejected,
	}
	tx, err := repo.db.Begin()
	if err != nil {
		processInfo.Reason = "internal server error. Failed to start transaction."
		return processInfo, []models.BatchOrderInventoryUpdate{}, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	queryOrder := `
        INSERT INTO orders (CustomerName, Notes)
        VALUES ($1, $2)
        RETURNING ID
    `

	notesJSON, err := json.Marshal(order.Notes)
	if err != nil {
		processInfo.Reason = "Notes field in invalid format. Must be json"
		return processInfo, []models.BatchOrderInventoryUpdate{}, fmt.Errorf("failed to marshal notes: %w", err)
	}

	var ID int
	err = tx.QueryRow(queryOrder, order.CustomerName, notesJSON).Scan(&ID)
	if err != nil {
		processInfo.Reason = "internal server error. Failed to scan ID"
		return processInfo, []models.BatchOrderInventoryUpdate{}, err
	}
	processInfo.OrderID = ID

	queryOrderItems := `
		INSERT INTO order_items (ProductID, Quantity, OrderID) VALUES
		($1, $2, $3)
		ON CONFLICT (OrderID, ProductID)
		DO UPDATE SET Quantity = order_items.Quantity + EXCLUDED.Quantity;
	`

	queryGetPrice := `
		SELECT price FROM menu_items WHERE id = $1
	`

	queryGetIngredients := `
		SELECT IngredientID, Quantity FROM menu_item_ingredients WHERE MenuID = $1
	`

	queryUpdateInventory := `
		UPDATE inventory SET Quantity = Quantity - $1 WHERE IngredientID = $2 AND Quantity >= $1
	`
	inventoryInfo := []models.BatchOrderInventoryUpdate{}
	for _, v := range order.Items {

		_, err = tx.Exec(queryOrderItems, v.ProductID, v.Quantity, ID)
		if err != nil {
			processInfo.Reason = "internal server error. " + err.Error()
			processInfo.Total = 0
			return processInfo, []models.BatchOrderInventoryUpdate{}, err
		}

		var price float64
		err = tx.QueryRow(queryGetPrice, v.ProductID).Scan(&price)
		if err != nil {
			processInfo.Reason = "internal server error." + err.Error()
			processInfo.Total = 0
			return processInfo, []models.BatchOrderInventoryUpdate{}, err
		}
		processInfo.Total += float64(v.Quantity) * price

		var ingredients []struct {
			IngredientID     int
			RequiredQuantity int
		}

		rows, err := tx.Query(queryGetIngredients, v.ProductID)
		if err != nil {
			processInfo.Reason = "internal server error. Failed to get ingredients."
			processInfo.Total = 0
			return processInfo, []models.BatchOrderInventoryUpdate{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var ingredient struct {
				IngredientID     int
				RequiredQuantity int
			}
			if err := rows.Scan(&ingredient.IngredientID, &ingredient.RequiredQuantity); err != nil {
				processInfo.Reason = "internal server error. Failed to scan ingredient."
				processInfo.Total = 0
				return processInfo, []models.BatchOrderInventoryUpdate{}, err
			}
			ingredients = append(ingredients, ingredient)
		}
		for _, ing := range ingredients {
			totalRequired := ing.RequiredQuantity * v.Quantity

			var availableQuantity int
			var InvName string

			err = tx.QueryRow("SELECT quantity, name FROM inventory WHERE IngredientID = $1", ing.IngredientID).Scan(&availableQuantity, &InvName)
			if err != nil {
				processInfo.Reason = fmt.Sprintf("internal server error. Failed to check inventory. ID=%d", ing.IngredientID)
				processInfo.Total = 0
				return processInfo, []models.BatchOrderInventoryUpdate{}, err
			}

			if availableQuantity < totalRequired {
				processInfo.Reason = fmt.Sprintf("insufficient_inventory. IngredientID: %d. Required: %d, Available: %d", ing.IngredientID, totalRequired, availableQuantity)
				processInfo.Total = 0
				return processInfo, []models.BatchOrderInventoryUpdate{}, fmt.Errorf("%s", processInfo.Reason)
			}

			_, err = tx.Exec(queryUpdateInventory, totalRequired, ing.IngredientID)
			if err != nil {
				processInfo.Reason = "internal server error. Failed to update inventory."
				processInfo.Total = 0
				return processInfo, []models.BatchOrderInventoryUpdate{}, err
			}

			InvInfo := models.BatchOrderInventoryUpdate{
				IngredientID:  ing.IngredientID,
				Name:          InvName,
				Quantity_used: totalRequired,
				Remaining:     availableQuantity - totalRequired,
			}
			inventoryInfo = append(inventoryInfo, InvInfo)
		}
	}

	err = tx.Commit()
	if err != nil {
		processInfo.Reason = "Internal server error. Error commiting transaction."
		processInfo.Total = 0
		return processInfo, inventoryInfo, err
	}
	processInfo.Status = models.StatusOrderAccepted
	processInfo.Reason = "OK"
	return processInfo, inventoryInfo, nil
}

func (repo *OrderRepository) GetAll() ([]models.Order, error) {
	query := `
	 SELECT ID, CustomerName, Status, Notes, CreatedAt
	 FROM orders`

	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order

	for rows.Next() {
		var order models.Order
		var notes []byte
		if err := rows.Scan(&order.ID, &order.CustomerName, &order.Status, &notes, &order.CreatedAt); err != nil {
			return nil, err
		}

		json.Unmarshal(notes, &order.Notes)

		items, err := getOrderItems(repo.db, order.ID)
		if err != nil {
			return nil, err
		}
		order.Items = items

		orders = append(orders, order)
	}

	return orders, nil
}

func (repo *OrderRepository) GetOrderByID(id int) (models.Order, error) {
	query := `
		SELECT ID, CustomerName, Status, Notes, CreatedAt
		FROM orders WHERE ID = $1`

	var order models.Order
	var notes []byte
	err := repo.db.QueryRow(query, id).Scan(&order.ID, &order.CustomerName, &order.Status, &notes, &order.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Order{}, models.ErrOrderNotFound
		}
		return models.Order{}, err
	}

	json.Unmarshal(notes, &order.Notes)

	items, err := getOrderItems(repo.db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Order{}, models.ErrOrderNotFound
		}
		return models.Order{}, err
	}
	order.Items = items
	return order, nil
}

func (repo *OrderRepository) SaveUpdatedOrder(updatedOrder models.Order, OrderID string) error {
	queryCheckStatus := `
	select Status from orders where ID = $1
	`
	var Status string
	err := repo.db.QueryRow(queryCheckStatus, OrderID).Scan(&Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.ErrOrderNotFound
		}
		return err
	}

	if Status == "closed" {
		return models.ErrOrderClosed
	}
	queryUpdateOrder := `
	update orders 
	set CustomerName = $1
	where ID = $2
	`
	_, err = repo.db.Query(queryUpdateOrder, updatedOrder.CustomerName, OrderID)
	if err != nil {
		return err
	}
	for _, v := range updatedOrder.Items {
		queryUpdateOrderItems := `
		update order_items set ProductID = $1, Quantity = $2 where OrderID = $3
		`
		_, err = repo.db.Query(queryUpdateOrderItems, v.ProductID, v.Quantity, OrderID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *OrderRepository) DeleteOrder(OrderID int) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var orderExists bool
	queryCheckOrder := `SELECT EXISTS(SELECT 1 FROM orders WHERE ID = $1)`
	err = tx.QueryRow(queryCheckOrder, OrderID).Scan(&orderExists)
	if err != nil {
		tx.Rollback()
		return err
	}

	if !orderExists {
		tx.Rollback()
		return fmt.Errorf("order not found", OrderID)
	}

	queryDeleteOrderItems := `
	DELETE FROM order_items
	WHERE OrderID = $1
	`
	_, err = tx.Exec(queryDeleteOrderItems, OrderID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete order items: %w", err)
	}

	queryDeleteOrder := `
	DELETE FROM orders
	WHERE ID = $1
	`
	_, err = tx.Exec(queryDeleteOrder, OrderID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete order: %w", err)
	}

	return tx.Commit()
}

func (repo *OrderRepository) CloseOrderRepo(id int) error {
	var status string
	queryCheckStatus := `
		SELECT status FROM orders WHERE ID = $1
	`
	err := repo.db.QueryRow(queryCheckStatus, id).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.ErrOrderNotFound
		}
		return err
	}

	if status == "closed" {
		return models.ErrOrderClosed
	}

	queryToClose := `
		UPDATE orders SET status = 'closed' WHERE ID = $1
	`
	_, err = repo.db.Exec(queryToClose, id)
	if err != nil {
		return err
	}

	return nil
}

func getOrderItems(db *sql.DB, orderID int) ([]models.OrderItem, error) {
	query := `
	 SELECT ProductID, Quantity
	 FROM order_items
	 WHERE OrderID = $1`

	rows, err := db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed request for order_items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItem

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("error scanning row in order_items: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (repo *OrderRepository) GetNumberOfItems(startDate, endDate time.Time) (map[string]int, error) {
	query := `
		SELECT
			m.Name,
			COALESCE(SUM(oi.Quantity), 0) AS total_quantity
		FROM
			menu_items m
		LEFT JOIN
			order_items oi ON m.ID = oi.ProductID
		LEFT JOIN
			orders o ON oi.OrderID = o.ID
		WHERE
			(o.CreatedAt BETWEEN $1 AND $2) AND o.Status = 'closed'
		GROUP BY
			m.Name
		ORDER BY
			total_quantity DESC;
	`
	rows, err := repo.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	result := make(map[string]int)

	for rows.Next() {
		var name string
		var quantity int
		if err := rows.Scan(&name, &quantity); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		result[name] = quantity
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating rows: %v", err)
	}

	return result, nil
}

func (repo *OrderRepository) GetEarliestDate() string {
	query := `
	SELECT min(CreatedAt) FROM orders
	`
	row, _ := repo.db.Query(query)
	var Date string
	row.Scan(&Date)
	return Date
}

func (repo *OrderRepository) OrderedItemsByDay(month, year int) (map[string]interface{}, error) {
	query := `
		SELECT EXTRACT(DAY FROM createdat) AS day, COUNT(*) AS order_count
		FROM orders
		WHERE EXTRACT(MONTH FROM createdat) = $1 `

	if year != -1 {
		query += `AND EXTRACT(YEAR FROM createdat) = $2 `
	}
	query += `GROUP BY day ORDER BY day`

	var rows *sql.Rows
	var err error

	if year != -1 {
		rows, err = repo.db.Query(query, month, year)
	} else {
		rows, err = repo.db.Query(query, month)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch ordered items by day: %w", err)
	}

	defer rows.Close()

	orderedItems := make([]map[string]int, 0)
	for rows.Next() {
		var day int
		var orderCount int
		if err := rows.Scan(&day, &orderCount); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		orderedItems = append(orderedItems, map[string]int{fmt.Sprintf("%d", day): orderCount})
	}

	var result map[string]interface{}
	if year != -1 {
		result = map[string]interface{}{
			"period":       "day",
			"month":        month,
			"year":         year,
			"orderedItems": orderedItems,
		}
	} else {
		result = map[string]interface{}{
			"period":       "day",
			"month":        month,
			"orderedItems": orderedItems,
		}
	}

	return result, nil
}

func (repo *OrderRepository) OrderedItemsByMonth(year int) (map[string]interface{}, error) {
	query := `
		SELECT 
			TO_CHAR(o.createdat, 'Month') AS month,
			COUNT(o.ID) AS total_orders
		FROM 
			orders o
		WHERE 
			EXTRACT(YEAR FROM o.createdat) = $1
			AND o.status = 'closed'
		GROUP BY 
			TO_CHAR(o.createdat, 'Month'), EXTRACT(MONTH FROM o.createdat)
		ORDER BY 
			EXTRACT(MONTH FROM o.createdat);
	`
	rows, err := repo.db.Query(query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderedItems := make(map[string]interface{})

	for rows.Next() {
		var month string
		var orderCount int

		if err := rows.Scan(&month, &orderCount); err != nil {
			return nil, err
		}

		month = strings.TrimSpace(month)

		orderedItems[month] = orderCount
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"period":       "month",
		"year":         year,
		"orderedItems": orderedItems,
	}

	return result, nil
}
