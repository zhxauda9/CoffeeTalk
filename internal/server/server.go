package server

import (
	"database/sql"
	"hot-coffee/internal/dal"
	"hot-coffee/internal/handler"
	"hot-coffee/internal/html_handler"
	"hot-coffee/internal/service"
	"log"
	"log/slog"
	"net/http"
)

func ServerLaunch(db *sql.DB, logger *slog.Logger) {
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./web"))
	fs2 := http.FileServer(http.Dir("./uploads"))

	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.Handle("/pictures", http.StripPrefix("/pictures", fs))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs2))
	mux.HandleFunc("/", html_handler.ServeHome)
	mux.HandleFunc("/admin", html_handler.ServeAdmin)
	mux.HandleFunc("/admin/menu", html_handler.ServeMenu)
	mux.HandleFunc("/admin/inventory", html_handler.ServeInventory)
	mux.HandleFunc("/admin/order", html_handler.ServeOrder)
	mux.HandleFunc("/menuCatalog", html_handler.ServeMenuCatalog)

	// - - - - - - - - - - - - - - INVENTORY - - - - - - - - - - - - - -
	inventoryRepo := dal.NewInventoryRepository(db)
	inventoryService := service.NewInventoryService(*inventoryRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService, logger)

	mux.HandleFunc("POST /inventory", inventoryHandler.PostInventory)
	mux.HandleFunc("GET /inventory", inventoryHandler.GetInventory)
	mux.HandleFunc("GET /inventory/{id}", inventoryHandler.GetInventoryItem)
	mux.HandleFunc("PUT /inventory/{id}", inventoryHandler.PutInventoryItem)
	mux.HandleFunc("DELETE /inventory/{id}", inventoryHandler.DeleteInventoryItem)
	mux.HandleFunc("GET /inventory/getLeftOvers", inventoryHandler.GetLeftOvers)

	// - - - - - - - - - - - - - - MENU - - - - - - - - - - - - - -

	menuRepo := dal.NewMenuRepository(db)
	menuService := service.NewMenuService(*menuRepo, *inventoryRepo)
	menuHandler := handler.NewMenuHandler(menuService, logger)

	mux.HandleFunc("POST /menu", menuHandler.PostMenu)
	mux.HandleFunc("GET /menu", menuHandler.GetMenu)
	mux.HandleFunc("GET /menu/{id}", menuHandler.GetMenuItem)
	mux.HandleFunc("GET /menu/{id}/image", menuHandler.GetMenuItemImage)
	mux.HandleFunc("PUT /menu/{id}", menuHandler.PutMenuItem)
	mux.HandleFunc("PUT /menu/{id}/image", menuHandler.PutMenuItemImage)
	mux.HandleFunc("DELETE /menu/{id}", menuHandler.DeleteMenuItem)
	mux.HandleFunc("DELETE /menu/{id}/image", menuHandler.DeleteMenuItemImage)

	// - - - - - - - - - - - - - - ORDER - - - - - - - - - - - - - -

	orderRepo := dal.NewOrderRepository(db)
	orderService := service.NewOrderService(*orderRepo, *menuRepo, *inventoryRepo)
	orderHandler := handler.NewOrderHandler(orderService, menuService, logger)

	mux.HandleFunc("POST /orders", orderHandler.PostOrder)
	mux.HandleFunc("GET /orders", orderHandler.GetOrders)
	mux.HandleFunc("GET /orders/{id}", orderHandler.GetOrder)
	mux.HandleFunc("PUT /orders/{id}", orderHandler.PutOrder)
	mux.HandleFunc("DELETE /orders/{id}", orderHandler.DeleteOrder)
	mux.HandleFunc("POST /orders/{id}/close", orderHandler.CloseOrder)
	mux.HandleFunc("GET /orders/numberOfOrderedItems", orderHandler.GetNumberOfOrdered)
	mux.HandleFunc("POST /orders/batch-process", orderHandler.BatchOrders)

	// - - - - - - - - - - - - - - REPORT - - - - - - - - - - - - - -
	aggregationRepo := dal.NewReportRespository(db)
	aggregationService := service.NewAggregationService(aggregationRepo)
	reportHandler := handler.NewAggregationHandler(orderService, aggregationService, logger)

	mux.HandleFunc("GET /reports/total-sales", reportHandler.TotalSalesHandler)
	mux.HandleFunc("GET /reports/popular-items", reportHandler.PopularItemsHandler)
	mux.HandleFunc("GET /reports/orderedItemsByPeriod", reportHandler.OrderByPeriod)
	mux.HandleFunc("GET /reports/search", reportHandler.SearchHandler)

	logger.Info("Server started", "Address", "http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
