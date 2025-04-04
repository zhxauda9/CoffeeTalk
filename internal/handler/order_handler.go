package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"hot-coffee/internal/error_handler"
	"hot-coffee/internal/service"
	"hot-coffee/models"
)

// OrderHandler handles HTTP requests related to orders, interacting with the OrderService and MenuService.
type OrderHandler struct {
	orderService *service.OrderService // Service for handling order-related operations.
	menuService  *service.MenuService  // Service for menu-related operations.
	logger       *slog.Logger          // Logger for logging various events.
}

// NewOrderHandler initializes and returns a new OrderHandler with the provided services and logger.
func NewOrderHandler(orderService *service.OrderService, menuService *service.MenuService, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{orderService: orderService, menuService: menuService, logger: logger}
}

// PostOrder handles the creation of a new order via HTTP POST request.
func (h *OrderHandler) PostOrder(w http.ResponseWriter, r *http.Request) {
	var NewOrder models.Order
	// Decode the request body into the Order model.
	err := json.NewDecoder(r.Body).Decode(&NewOrder)
	if err != nil {
		h.logger.Error("Could not decode request json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not decode request json data", http.StatusBadRequest)
		return
	}

	// Validate each item in the order.
	for _, OrderItem := range NewOrder.Items {
		// Check if the product exists in the menu.
		if err = h.menuService.MenuCheckByID(OrderItem.ProductID, true); err != nil {
			h.logger.Error("Requested order item does not exist in menu", "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, "Requested order item does not exist in menu", http.StatusBadRequest)
			return
		}
		// Validate ingredient availability based on quantity.
		if err = h.menuService.IngredientsCheckByID(OrderItem.ProductID, OrderItem.Quantity); err != nil {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Add the order using the order service.
	_, _, err = h.orderService.AddOrder(NewOrder)
	if err != nil {
		if err.Error() == "something wrong with your requested order" {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, "Something wrong when adding new order", http.StatusInternalServerError)
			return
		}
	}

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.WriteHeader(http.StatusCreated)
}

// GetOrders handles the retrieval of all orders via HTTP GET request.
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	Orders, err := h.orderService.GetAllOrders()
	if err != nil {
		h.logger.Error("Can not read order data from server", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Can not read order data from server", http.StatusInternalServerError)
		return
	}
	// Marshal the orders to JSON.
	jsonData, err := json.MarshalIndent(Orders, "", "    ")
	if err != nil {
		h.logger.Error("Can not convert order data to json", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Can not convert order data to json", http.StatusInternalServerError)
		return
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonData)
}

// GetOrder handles the retrieval of a single order by ID via HTTP GET request.
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		error_handler.Error(w, "The id should be positive integer", http.StatusBadRequest)
		h.logger.Error("The id should be positive integer", "method", r.Method, "url", r.URL)
		return
	}
	RequestedOrder, err := h.orderService.GetOrder(ID)
	if err != nil {
		if err.Error() == "order not found" {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, err.Error(), http.StatusNotFound)
			return
		} else {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// Marshal the requested order to JSON.
	jsonData, err := json.MarshalIndent(RequestedOrder, "", "    ")
	if err != nil {
		h.logger.Error("Can not convert order data to json", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Can not convert order data to json", http.StatusInternalServerError)
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonData)
}

// PutOrder handles updating an existing order via HTTP PUT request.
func (h *OrderHandler) PutOrder(w http.ResponseWriter, r *http.Request) {
	var RequestedOrder models.Order
	// Decode the request body into the Order model.
	err := json.NewDecoder(r.Body).Decode(&RequestedOrder)
	if err != nil {
		h.logger.Error("Could not decode request json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not decode request json data", http.StatusBadRequest)
		return
	}

	// Validate each item in the updated order.
	for _, OrderItem := range RequestedOrder.Items {
		// Check if the product exists in the menu.
		if err = h.menuService.MenuCheckByID(OrderItem.ProductID, true); err != nil {
			h.logger.Error("Updated order item does not exist in menu", "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, "Updated order item does not exist in menu", http.StatusBadRequest)
			return
		}
		// Validate ingredient availability based on quantity.
		if err = h.menuService.IngredientsCheckByID(OrderItem.ProductID, OrderItem.Quantity); err != nil {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	// Update the order in the service.
	err = h.orderService.UpdateOrder(RequestedOrder, r.PathValue("id"))
	if err != nil {
		if err.Error() == "could not update the order because it is already closed" || err.Error() == "something wrong with your updated order" || err.Error() == "the requested order does not exist" {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.WriteHeader(200)
}

// DeleteOrder handles the deletion of an order via HTTP DELETE request.
func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		error_handler.Error(w, "The id should be positive integer", http.StatusBadRequest)
		h.logger.Error("The id should be positive integer", "method", r.Method, "url", r.URL)
		return
	}
	// Delete the order by ID using the order service.
	err = h.orderService.DeleteOrderByID(ID)
	if err != nil {
		if err.Error() == "order not found" {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, err.Error(), http.StatusNotFound)
			return
		} else {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, "Error updating orders database", http.StatusInternalServerError)
			return
		}
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.WriteHeader(204)
}

// CloseOrder handles the closing of an order via HTTP request.
func (h *OrderHandler) CloseOrder(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	ID, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Order id must be integer", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Order id must be integer", http.StatusBadRequest)
		return
	}

	// Close the order using the order service.
	err = h.orderService.CloseOrder(ID)
	if err != nil {
		h.logger.Error("Error closing order", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

// GetNumberOfOrdered handles the retrieval of the number of ordered items within a specific date range.
func (h *OrderHandler) GetNumberOfOrdered(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	if startDate == "" {
		startDate = "1970-01-01"
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	// Get the number of ordered items using the order service.
	items, err := h.orderService.GetNumberOfItems(startDate, endDate)
	if err != nil {
		h.logger.Error(err.Error(), "query", r.URL.Query, "error", err)
		error_handler.Error(w, fmt.Sprintf("Error getting number of ordered items. Error:%v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		h.logger.Error("Could not encode json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not encode request json data", http.StatusInternalServerError)
	}
}

// BatchOrders handles the batch processing of multiple orders via HTTP request.
func (h *OrderHandler) BatchOrders(w http.ResponseWriter, r *http.Request) {
	request := struct {
		Orders []models.Order `json:"orders"`
	}{}
	// Decode the request body into the batch of orders.
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.logger.Error("Could not decode request json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not decode request json data", http.StatusBadRequest)
		return
	}

	// Process the batch of orders using the order service.
	ordersReport, err := h.orderService.BulkOrders(request.Orders)
	if err != nil {
		h.logger.Error("Error processing orders", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Error processing orders. "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ordersReport); err != nil {
		h.logger.Error("Error encoding response", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
