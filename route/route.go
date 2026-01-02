package route

import (
	"reserv-service/database"
	handler "reserv-service/handle"
	"reserv-service/repository"
	"reserv-service/service"

	"github.com/gin-gonic/gin"
)

// InitRoute создает и настраивает роутер
func InitRoute() *gin.Engine {
	r := gin.Default()

	// Получаем подключение к БД
	db := database.GetDB()

	// Инициализируем слои приложения
	reserveRepo := repository.NewReserveRepository(db)
	reserveService := service.NewReserveService(reserveRepo)
	reserveHandler := handler.NewReserveHandler(reserveService)

	// Health check
	r.GET("/reserv/health", func(c *gin.Context) {
		// Проверяем подключение к БД
		if db == nil {
			c.JSON(503, gin.H{
				"status":  "error",
				"message": "Database not connected",
			})
			return
		}

		// Проверяем подключение через ping
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(503, gin.H{
				"status":  "error",
				"message": "Failed to get DB connection",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(503, gin.H{
				"status":  "error",
				"message": "Database ping failed",
			})
			return
		}

		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "reservation-service",
			"version": "1.0.0",
		})
	})

	// API v1 группировка
	apiV1 := r.Group("/api/v1")
	{
		// Reserve endpoints
		reserves := apiV1.Group("/reserves")
		{
			reserves.POST("/", reserveHandler.PostReserve)
			reserves.GET("/user/:user_id/item/:item_id", reserveHandler.GetReserves)
			reserves.DELETE("/:id", reserveHandler.CancelReserve)
		}

		// Legacy endpoints (для обратной совместимости)
		apiV1.POST("/reserv/reserve", reserveHandler.PostReserve)
		apiV1.GET("/reserv/user/:user_id/item/:item_id", reserveHandler.GetReserves)
		apiV1.DELETE("/reserv/:id", reserveHandler.CancelReserve)
	}

	// Простой endpoint для теста
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Reservation Service API",
			"endpoints": []gin.H{
				{"method": "GET", "path": "/reserv/health", "description": "Health check"},
				{"method": "POST", "path": "/api/v1/reserves", "description": "Create reservation"},
				{"method": "GET", "path": "/api/v1/reserves/user/:user_id/item/:item_id", "description": "Get user reserves"},
				{"method": "DELETE", "path": "/api/v1/reserves/:id", "description": "Cancel reservation"},
			},
		})
	})

	return r
}
