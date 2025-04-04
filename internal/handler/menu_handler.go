package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"hot-coffee/internal/error_handler"
	"hot-coffee/internal/service"
	"hot-coffee/models"
)

// MenuHandler struct handles HTTP requests related to the menu and menu items.
type MenuHandler struct {
	menuService *service.MenuService // Service to interact with the menu data
	logger      *slog.Logger         // Logger for logging errors and info
}

// NewMenuHandler is a constructor to initialize MenuHandler with MenuService and Logger.
func NewMenuHandler(menuService *service.MenuService, logger *slog.Logger) *MenuHandler {
	return &MenuHandler{menuService: menuService, logger: logger}
}

// PostMenu adds a new menu item with an optional image uploaded as part of a multipart form.
func (h *MenuHandler) PostMenu(w http.ResponseWriter, r *http.Request) {
	var newItem models.MenuItem
	imagePath := "uploads/default.jpg" // Default image path
	contentType := r.Header.Get("Content-Type")

	// Check if the request content is JSON or multipart (form data)
	if contentType == "application/json" {
		// If JSON, decode the request body into a MenuItem
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&newItem); err != nil {
			error_handler.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
	} else {
		// If multipart, handle file upload and form fields
		err := r.ParseMultipartForm(10 << 20) // 10MB limit for file size
		if err != nil {
			error_handler.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Handle image file upload
		file, handler, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			ext := filepath.Ext(handler.Filename)
			imagePath = fmt.Sprintf("uploads/menu-%d%s", time.Now().UnixNano(), ext)

			// Save the image file to disk
			outFile, err := os.Create(imagePath)
			if err != nil {
				error_handler.Error(w, "Could not save image", http.StatusInternalServerError)
				return
			}
			defer outFile.Close()
			io.Copy(outFile, file)
		}

		// Parse the rest of the form data
		newItem.Name = r.FormValue("name")
		newItem.Description = r.FormValue("description")
		newItem.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
		json.Unmarshal([]byte(r.FormValue("ingredients")), &newItem.Ingredients)

		// Validate that all required fields are provided
		if newItem.Name == "" || newItem.Description == "" || newItem.Price == 0 {
			error_handler.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}
	}

	// Assign the image path to the new item
	newItem.Image = imagePath

	// Check for various conditions like valid menu and ingredient checks
	if err := h.menuService.CheckNewMenu(newItem); err != nil {
		error_handler.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.menuService.MenuCheckByID(newItem.ID, false); err != nil {
		error_handler.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.menuService.IngredientsCheckForNewItem(newItem); err != nil {
		error_handler.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add the menu item
	if err := h.menuService.AddMenuItem(newItem); err != nil {
		error_handler.Error(w, "Could not add menu item", http.StatusInternalServerError)
		return
	}

	// Respond with status created (201)
	w.WriteHeader(http.StatusCreated)
}

// DeleteMenuItemImage resets the image of a menu item to a default image.
func (h *MenuHandler) DeleteMenuItemImage(w http.ResponseWriter, r *http.Request) {
	// Extract ID from the URL
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		error_handler.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Update image to default
	defaultImage := "uploads/default.jpg"
	err = h.menuService.UpdateMenuItemImage(id, defaultImage)
	if err != nil {
		error_handler.Error(w, "Could not reset image", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	h.logger.Info("Image reset to default", "id", id)
}

// GetMenu retrieves all menu items and returns them as a JSON array.
func (h *MenuHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	MenuItems, err := h.menuService.GetMenuItems()
	if err != nil {
		h.logger.Error("Could not read menu database", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not read menu database", http.StatusInternalServerError)
		return
	}
	jsonData, err := json.MarshalIndent(MenuItems, "", "    ")
	if err != nil {
		h.logger.Error("Could not read menu database", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not read menu items", http.StatusInternalServerError)
		return
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// GetMenuItem retrieves a single menu item based on its ID.
func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Menu id must be integer", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Menu id must be integer", http.StatusBadRequest)
		return
	}

	// Retrieve the menu item by ID
	MenuItem, err := h.menuService.GetMenuItem(id)
	if err != nil {
		// Handle item not found case
		if err.Error() == "could not find menu item by the given id" {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, err.Error(), http.StatusNotFound)
			return
		} else {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			error_handler.Error(w, "Could not read menu database", http.StatusInternalServerError)
			return
		}
	}
	jsonData, err := json.MarshalIndent(MenuItem, "", "    ")
	if err != nil {
		h.logger.Error("Could not convert Menu Items to jsondata", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not send menu item", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

// PutMenuItem updates an existing menu item by its ID.
func (h *MenuHandler) PutMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Menu id must be integer", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Menu id must be integer", http.StatusBadRequest)
		return
	}

	var requestedMenuItem models.MenuItem
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestedMenuItem); err != nil {
		h.logger.Error("Could not parse JSON", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	requestedMenuItem.ID = id

	// Check if the menu item exists and validate
	if err = h.menuService.MenuCheckByID(id, true); err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate and check ingredients for the new item
	if err = h.menuService.CheckNewMenu(requestedMenuItem); err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = h.menuService.IngredientsCheckForNewItem(requestedMenuItem); err != nil {
		h.logger.Error(err.Error(), "method", r.Method, "url", r.URL)
		error_handler.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Use default image if no image path is provided
	if requestedMenuItem.Image == "" {
		requestedMenuItem.Image = "uploads/default.jpg"
	}

	// Update the menu item in the service/database
	if err = h.menuService.UpdateMenuItem(requestedMenuItem); err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not update menu database", http.StatusInternalServerError)
		return
	}

	// Respond with status OK (200)
	w.WriteHeader(http.StatusOK)
	h.logger.Info("Menu item updated successfully.", "method", r.Method, "url", r.URL)
}

// PutMenuItemImage updates the image of a menu item.
func (h *MenuHandler) PutMenuItemImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		error_handler.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Parse form data and handle file upload
	err = r.ParseMultipartForm(10 << 20) // 10MB limit for file size
	if err != nil {
		error_handler.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		error_handler.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate the file format
	ext := filepath.Ext(handler.Filename)
	if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
		error_handler.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}

	// Save the image file
	imagePath := fmt.Sprintf("uploads/menu-%d%s", id, ext)

	outFile, err := os.Create(imagePath)
	if err != nil {
		error_handler.Error(w, "Could not save image", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, file)
	if err != nil {
		error_handler.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	// Update the image path in the database
	err = h.menuService.UpdateMenuItemImage(id, imagePath)
	if err != nil {
		error_handler.Error(w, "Could not update image path in database", http.StatusInternalServerError)
		return
	}

	// Respond with status OK (200)
	w.WriteHeader(http.StatusOK)
}

// DeleteMenuItem deletes a menu item by its ID.
func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Menu id must be integer", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Menu id must be integer", http.StatusBadRequest)
		return
	}

	// Delete the menu item from the database
	err = h.menuService.DeleteMenuItem(id)
	if err != nil {
		h.logger.Error("Could not delete menu item", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not delete menu item", http.StatusInternalServerError)
		return
	}
	// Respond with no content (204)
	w.WriteHeader(204)
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

// GetMenuItemImage serves the image for a specific menu item.
func (h *MenuHandler) GetMenuItemImage(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from URL
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		error_handler.Error(w, "Menu id must be integer", http.StatusBadRequest)
		return
	}
	// Retrieve the menu item and serve its image
	menuItem, err := h.menuService.GetMenuItem(id)
	if err != nil {
		error_handler.Error(w, "Menu item not found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, menuItem.Image)
}
