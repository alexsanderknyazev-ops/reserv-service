package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reserv-service/models"
	"reserv-service/service"
	"time"

	"github.com/gin-gonic/gin"
)

type ReserveHandler struct {
	reserveService *service.ReserveService
}

func NewReserveHandler(reserveService *service.ReserveService) *ReserveHandler {
	return &ReserveHandler{reserveService: reserveService}
}

// PostReserve обрабатывает POST запрос на создание резерва
func (h *ReserveHandler) PostReserve(c *gin.Context) {
	var req models.ReserveRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Создаем резерв
	reserve, err := h.reserveService.CreateReserve(c.Request.Context(), req)
	if err != nil {
		log.Printf("Error creating reserve: %v", err)

		statusCode := http.StatusInternalServerError
		if err == service.ErrItemAlreadyReserved {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Формируем ответ
	response := models.ReserveResponse{
		ID:        reserve.ID,
		IdItem:    reserve.IdItem,
		IdUser:    reserve.IdUser,
		ItemType:  string(reserve.ItemType),
		Status:    reserve.Status,
		ExpiresAt: reserve.ExpiresAt,
		CreatedAt: reserve.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetReserves возвращает все резервы пользователя
func (h *ReserveHandler) GetReserves(c *gin.Context) {
	userID := c.Param("user_id")

	// Конвертируем userID в int64
	var idUser int64
	if _, err := fmt.Sscanf(userID, "%d", &idUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	reserves, err := h.reserveService.GetUserReserves(c.Request.Context(), idUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, reserves)
}

// CancelReserve отменяет резерв
func (h *ReserveHandler) CancelReserve(c *gin.Context) {
	reserveID := c.Param("id")

	var id int64
	if _, err := fmt.Sscanf(reserveID, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid reserve ID",
		})
		return
	}

	err := h.reserveService.CancelReserve(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrReserveNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Reserve not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reserve cancelled successfully",
	})
}

// Чистый HTTP handler (без Gin)
func PostReserveHTTP(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// Читаем тело запроса
	var req models.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация
	if req.IdUser <= 0 || req.IdItem <= 0 {
		http.Error(w, "Invalid user or item ID", http.StatusBadRequest)
		return
	}

	// Здесь будет логика создания резерва через сервис
	// Пока заглушка
	reserve := models.Reserve{
		IdItem:    req.IdItem,
		IdUser:    req.IdUser,
		ItemType:  req.ItemType,
		Status:    "pending",
		ExpiresAt: time.Now().Add(30 * time.Minute),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reserve)
}
