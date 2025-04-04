package html_handler

import (
	"net/http"
	"path/filepath"
)

func ServeHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "home.html"))
}

func ServeAdmin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "admin.html"))
}

func ServeMenu(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "menu.html"))
}

func ServeInventory(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "inventory.html"))
}

func ServeOrder(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "order.html"))
}

func ServeMenuCatalog(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "menuCatalog.html"))
}
