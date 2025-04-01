package handler

import (
	"encoding/json"
	"fmt"
	"hot-coffee/internal/error_handler"
	"hot-coffee/internal/service"
	"hot-coffee/models"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

type MenuHandler struct {
	menuService *service.MenuService
	logger      *slog.Logger
}

func NewMenuHandler(menuService *service.MenuService, logger *slog.Logger) *MenuHandler {
	return &MenuHandler{menuService: menuService, logger: logger}
}

func (h *MenuHandler) PostMenu(w http.ResponseWriter, r *http.Request) {
	var newItem models.MenuItem
	imagePath := "uploads/default.jpg"
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&newItem); err != nil {
			error_handler.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
	} else {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			error_handler.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("image")
		if err == nil { // Если изображение передано, сохраняем его
			defer file.Close()
			imagePath = "uploads/" + handler.Filename
			outFile, err := os.Create(imagePath)
			if err != nil {
				error_handler.Error(w, "Could not save image", http.StatusInternalServerError)
				return
			}
			defer outFile.Close()
			io.Copy(outFile, file)
		}

		newItem.Name = r.FormValue("name")
		newItem.Description = r.FormValue("description")
		newItem.Price, _ = strconv.ParseFloat(r.FormValue("price"), 64)
		json.Unmarshal([]byte(r.FormValue("ingredients")), &newItem.Ingredients)
	}

	newItem.Image = imagePath

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
	if err := h.menuService.AddMenuItem(newItem); err != nil {
		error_handler.Error(w, "Could not add menu item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *MenuHandler) DeleteMenuItemImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		error_handler.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	defaultImage := "uploads/default.jpg"
	err = h.menuService.UpdateMenuItemImage(id, defaultImage)
	if err != nil {
		error_handler.Error(w, "Could not reset image", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.logger.Info("Image reset to default", "id", id)
}

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

func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Menu id must be integer", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Menu id must be integer", http.StatusBadRequest)
		return
	}

	MenuItem, err := h.menuService.GetMenuItem(id)
	if err != nil {
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

	if err = h.menuService.MenuCheckByID(id, true); err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	if requestedMenuItem.Image == "" {
		requestedMenuItem.Image = "uploads/default.jpg"
	}

	if err = h.menuService.UpdateMenuItem(requestedMenuItem); err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not update menu database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.logger.Info("Menu item updated successfully.", "method", r.Method, "url", r.URL)
}

func (h *MenuHandler) PutMenuItemImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		error_handler.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	imagePath := fmt.Sprintf("uploads/menu_%d.jpg", id)

	outFile, err := os.Create(imagePath)
	if err != nil {
		error_handler.Error(w, "Could not save image", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, r.Body)
	if err != nil {
		error_handler.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	if err = h.menuService.UpdateMenuItemImage(id, imagePath); err != nil {
		error_handler.Error(w, "Could not update image", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	h.logger.Info("Image uploaded successfully", "id", id)
}

func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Menu id must be integer", "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Menu id must be integer", http.StatusBadRequest)
		return
	}

	err = h.menuService.DeleteMenuItem(id)
	if err != nil {
		h.logger.Error("Could not delete menu item", "error", err, "method", r.Method, "url", r.URL)
		error_handler.Error(w, "Could not delete menu item", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(204)
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *MenuHandler) GetMenuItemImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		error_handler.Error(w, "Menu id must be integer", http.StatusBadRequest)
		return
	}
	menuItem, err := h.menuService.GetMenuItem(id)
	if err != nil {
		error_handler.Error(w, "Menu item not found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, menuItem.Image)
}
