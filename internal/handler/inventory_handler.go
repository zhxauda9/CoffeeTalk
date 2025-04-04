package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"hot-coffee/internal/error_handler"
	"hot-coffee/internal/service"
	"hot-coffee/models"
)

// InventoryHandler handles HTTP requests related to inventory items.
type InventoryHandler struct {
	inventoryService *service.InventoryService
	logger           *slog.Logger
}

// NewInventoryHandler creates a new InventoryHandler instance.
func NewInventoryHandler(inventoryService *service.InventoryService, logger *slog.Logger) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService, logger: logger}
}

// PostInventory handles creating a new inventory item.
func (h *InventoryHandler) PostInventory(w http.ResponseWriter, r *http.Request) {
	var newItem models.InventoryItem
	// Decode the incoming JSON data
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		h.logger.Error("Could not decode json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not decode request json data", http.StatusInternalServerError)
		return
	}

	// Validate required fields (Name, Unit, Quantity)
	if newItem.Name == "" || newItem.Unit == "" || newItem.Quantity <= 0 {
		h.logger.Error("Some fields are empty, equal or less than zero", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Some fields are empty, equal or less than zero", http.StatusBadRequest)
		return
	}

	// Add the new inventory item to the database
	err = h.inventoryService.AddInventoryItem(newItem)
	if err != nil {
		h.logger.Error("Could not add new inventory item", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not add new inventory item Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Successfully added item
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.WriteHeader(http.StatusCreated)
}

// GetInventory retrieves all inventory items.
func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	inventoryItems, err := h.inventoryService.GetAllInventoryItems()
	if err != nil {
		h.logger.Error("Could not get inventory items", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not get inventory items", http.StatusInternalServerError)
		return
	}

	// Marshal inventory items to JSON
	jsonData, err := json.Marshal(inventoryItems)
	if err != nil {
		h.logger.Error("Could not encode json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not encode request json data", http.StatusInternalServerError)
		return
	}

	// Respond with the inventory items in JSON format
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

// GetInventoryItem retrieves a single inventory item by ID.
func (h *InventoryHandler) GetInventoryItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	// Convert ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Inventory id must be integer", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Inventory id must be integer", http.StatusBadRequest)
		return
	}

	// Check if the inventory item exists
	if !h.inventoryService.Exists(id) {
		h.logger.Error("Inventory item does not exists", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Inventory item does not exists", http.StatusBadRequest)
		return
	}

	// Retrieve the inventory item from the service
	inventoryItem, err := h.inventoryService.GetItem(id)
	if err != nil {
		h.logger.Error("Could not get inventory item", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not get inventory item", http.StatusInternalServerError)
		return
	}

	// Marshal the inventory item to JSON
	jsonData, err := json.Marshal(inventoryItem)
	if err != nil {
		h.logger.Error("Could not encode json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not encode json data", http.StatusInternalServerError)
		return
	}
	// Respond with the inventory item
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

// PutInventoryItem updates an existing inventory item by ID.
func (h *InventoryHandler) PutInventoryItem(w http.ResponseWriter, r *http.Request) {
	var newItem models.InventoryItem
	// Decode the incoming JSON data
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		h.logger.Error("Could not decode json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not decode request json data", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if newItem.Name == "" || newItem.Unit == "" || newItem.Quantity <= 0 {
		h.logger.Error("Some fields are empty", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Some fields are empty, equal or less than zero", http.StatusBadRequest)
		return
	}

	// Get the inventory item ID from the URL path
	idStr := r.PathValue("id")

	// Convert ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Inventory id must be integer", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Inventory id must be integer", http.StatusBadRequest)
		return
	}

	// Check if the inventory item exists
	if !h.inventoryService.Exists(id) {
		h.logger.Error("Inventory item does not exists", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Inventory item does not exists", http.StatusBadRequest)
		return
	}

	// Update the inventory item
	err = h.inventoryService.UpdateItem(id, newItem)
	if err != nil {
		h.logger.Error("Error updating inventory item", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Error updating inventory item Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

// DeleteInventoryItem deletes an inventory item by ID.
func (h *InventoryHandler) DeleteInventoryItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	// Convert ID to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Inventory id must be integer", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Inventory id must be integer", http.StatusBadRequest)
		return
	}

	// Check if the inventory item exists
	if !h.inventoryService.Exists(id) {
		h.logger.Error("Inventory item does not exists", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Inventory item does not exists", http.StatusBadRequest)
		return
	}

	// Delete the inventory item
	err = h.inventoryService.DeleteItem(id)
	if err != nil {
		h.logger.Error("Could not delete inventory item", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not delete inventory item", http.StatusInternalServerError)
		return
	}

	// Successfully deleted the item
	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

// GetLeftOvers retrieves items from inventory that are considered leftovers.
func (h *InventoryHandler) GetLeftOvers(w http.ResponseWriter, r *http.Request) {
	// Retrieve query parameters
	ParamSortBy := r.URL.Query().Get("sortBy")
	ParamPage := r.URL.Query().Get("page")
	ParamPageSize := r.URL.Query().Get("pageSize")

	// Fetch leftover inventory items
	resp, err := h.inventoryService.GetLeftOvers(ParamSortBy, ParamPage, ParamPageSize)
	if err != nil {
		error_handler.Error(w, fmt.Sprintf("Error %v", err), http.StatusBadRequest)
		return
	}

	// Respond with leftover items in JSON format
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Could not encode json data", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not encode request json data", http.StatusInternalServerError)
	}
}
