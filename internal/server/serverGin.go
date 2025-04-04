package server

// import (
// 	"database/sql"
// 	"hot-coffee/internal/dal"
// 	"hot-coffee/internal/handler"
// 	"hot-coffee/internal/html_handler"
// 	"hot-coffee/internal/service"
// 	"log"
// 	"log/slog"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// func ServerGINLaunch(db *sql.DB, logger *slog.Logger) {
// 	router := gin.Default()

// 	fs := http.FileServer(http.Dir("./web"))
// 	fs2 := http.FileServer(http.Dir("./uploads"))

// 	router.Handle("/static/", http.StripPrefix("/static/", fs))
// 	router.Handle("/pictures", http.StripPrefix("/pictures", fs))
// 	router.Handle("/uploads/", http.StripPrefix("/uploads/", fs2))
// 	router.GET("/", html_handler.ServeHome)
// 	router.GET("/admin", html_handler.ServeAdmin)
// 	router.GET("/admin/menu", html_handler.ServeMenu)
// 	router.GET("/admin/inventory", html_handler.ServeInventory)
// 	router.GET("/admin/order", html_handler.ServeOrder)
// 	router.GET("/menuCatalog", html_handler.ServeMenuCatalog)

// 	// - - - - - - - - - - - - - - INVENTORY - - - - - - - - - - - - - -
// 	inventoryRepo := dal.NewInventoryRepository(db)
// 	inventoryService := service.NewInventoryService(*inventoryRepo)
// 	inventoryHandler := handler.NewInventoryHandler(inventoryService, logger)

// 	router.POST("/inventory", inventoryHandler.PostInventory)
// 	router.GET("/inventory", inventoryHandler.GetInventory)
// 	router.GET("/inventory/{id}", inventoryHandler.GetInventoryItem)
// 	router.PUT("/inventory/{id}", inventoryHandler.PutInventoryItem)
// 	router.DELETE("/inventory/{id}", inventoryHandler.DeleteInventoryItem)
// 	router.GET("/inventory/getLeftOvers", inventoryHandler.GetLeftOvers)

// 	// - - - - - - - - - - - - - - MENU - - - - - - - - - - - - - -

// 	menuRepo := dal.NewMenuRepository(db)
// 	menuService := service.NewMenuService(*menuRepo, *inventoryRepo)
// 	menuHandler := handler.NewMenuHandler(menuService, logger)

// 	router.POST("/menu", menuHandler.PostMenu)
// 	router.GET("/menu", menuHandler.GetMenu)
// 	router.GET("/menu/{id}", menuHandler.GetMenuItem)
// 	router.GET("/menu/{id}/image", menuHandler.GetMenuItemImage)
// 	router.PUT("/menu/{id}", menuHandler.PutMenuItem)
// 	router.PUT("/menu/{id}/image", menuHandler.PutMenuItemImage)
// 	router.DELETE("/menu/{id}", menuHandler.DeleteMenuItem)
// 	router.DELETE("/menu/{id}/image", menuHandler.DeleteMenuItemImage)

// 	// - - - - - - - - - - - - - - ORDER - - - - - - - - - - - - - -

// 	orderRepo := dal.NewOrderRepository(db)
// 	orderService := service.NewOrderService(*orderRepo, *menuRepo, *inventoryRepo)
// 	orderHandler := handler.NewOrderHandler(orderService, menuService, logger)

// 	router.POST("/orders", orderHandler.PostOrder)
// 	router.GET("/orders", orderHandler.GetOrders)
// 	router.GET("/orders/{id}", orderHandler.GetOrder)
// 	router.PUT("PUT /orders/{id}", orderHandler.PutOrder)
// 	router.DELETE("DELETE /orders/{id}", orderHandler.DeleteOrder)
// 	router.POST("POST /orders/{id}/close", orderHandler.CloseOrder)
// 	router.GET("GET /orders/numberOfOrderedItems", orderHandler.GetNumberOfOrdered)
// 	router.POST("POST /orders/batch-process", orderHandler.BatchOrders)

// 	// - - - - - - - - - - - - - - REPORT - - - - - - - - - - - - - -
// 	aggregationRepo := dal.NewReportRespository(db)
// 	aggregationService := service.NewAggregationService(aggregationRepo)
// 	reportHandler := handler.NewAggregationHandler(orderService, aggregationService, logger)

// 	router.GET("/reports/total-sales", reportHandler.TotalSalesHandler)
// 	router.GET("/reports/popular-items", reportHandler.PopularItemsHandler)
// 	router.GET("/reports/orderedItemsByPeriod", reportHandler.OrderByPeriod)
// 	router.GET("/reports/search", reportHandler.SearchHandler)

// 	logger.Info("Server started", "Address", "http://localhost:8080/")
// 	log.Fatal(http.ListenAndServe(":8080", router))
// }
