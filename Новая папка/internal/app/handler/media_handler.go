package handler

import (
	"net/http"
	"space_astrophysics/internal/app/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// =========================================================
// 🖼️ MEDIA (фото и видео для планет)
// =========================================================

// GetPlanetMedia получает все медиа файлы для планеты, отсортированные по ID
// @Summary        Получить медиа файлы планеты
// @Description    Возвращает все фото и видео для указанной планеты, отсортированные по ID (порядок в карусели)
// @Tags           media
// @Produce        json
// @Param          id   path    int  true    "ID планеты"
// @Success        200         {array} models.Media
// @Failure        404         {object} map[string]string
// @Router         /api/media/planet/{id} [get]
func (h *Handler) GetPlanetMedia(ctx *gin.Context) {
	planetID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID планеты"})
		return
	}

	// Проверяем существование планеты
	if _, err := h.Repo.GetPlanetByID(planetID); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Планета не найдена"})
		return
	}

	// Получаем медиа файлы
	var mediaFiles []models.Media
	if err := h.Repo.DB.Where("planet_id = ?", planetID).Order("id ASC").Find(&mediaFiles).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения медиа файлов"})
		return
	}

	ctx.JSON(http.StatusOK, mediaFiles)
}

// AddPlanetMedia добавляет медиа файл к планете
// @Summary        Добавить медиа файл
// @Description    Добавляет фото или видео к планете
// @Tags           media
// @Accept         json
// @Produce        json
// @Param          id   path    int          true    "ID планеты"
// @Param          media       body    models.Media true    "Данные медиа файла"
// @Success        201         {object} models.Media
// @Failure        400         {object} map[string]string
// @Router         /api/media/planet/{id} [post]
// @Security       BearerAuth
func (h *Handler) AddPlanetMedia(ctx *gin.Context) {
	planetID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID планеты"})
		return
	}

	// Проверяем существование планеты
	if _, err := h.Repo.GetPlanetByID(planetID); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Планета не найдена"})
		return
	}

	var media models.Media
	if err := ctx.ShouldBindJSON(&media); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON: " + err.Error()})
		return
	}

	// Проверяем тип файла
	if media.FileType != "image" && media.FileType != "video" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Тип файла должен быть 'image' или 'video'"})
		return
	}

	// Устанавливаем ID планеты
	media.PlanetID = planetID

	// Сохраняем в БД
	if err := h.Repo.DB.Create(&media).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения медиа файла: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, media)
}

// DeleteMedia удаляет медиа файл по ID
// @Summary        Удалить медиа файл
// @Description    Удаляет фото или видео по ID (только для модераторов)
// @Tags           media
// @Produce        json
// @Param          id  path    int  true    "ID медиа файла"
// @Success        200 {object} map[string]string
// @Failure        404 {object} map[string]string
// @Router         /api/media/{id} [delete]
// @Security       BearerAuth
func (h *Handler) DeleteMedia(ctx *gin.Context) {
	mediaID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID медиа файла"})
		return
	}

	// Проверяем роль (только модератор может удалять)
	role := ctx.GetString("role")
	if role != "mission_control" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Только модератор может удалять медиа файлы"})
		return
	}

	// Проверяем существование медиа файла
	var media models.Media
	if err := h.Repo.DB.First(&media, mediaID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Медиа файл не найден"})
		return
	}

	// Удаляем из БД
	if err := h.Repo.DB.Delete(&media).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления медиа файла"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Медиа файл успешно удален",
		"media_id":  mediaID,
		"planet_id": media.PlanetID,
		"file_url":  media.FileURL,
	})
}

// GetMediaByID получает медиа файл по ID
// @Summary        Получить медиа файл по ID
// @Description    Возвращает информацию о медиа файле
// @Tags           media
// @Produce        json
// @Param          id  path    int  true    "ID медиа файла"
// @Success        200 {object} models.Media
// @Failure        404 {object} map[string]string
// @Router         /api/media/{id} [get]
func (h *Handler) GetMediaByID(ctx *gin.Context) {
	mediaID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID медиа файла"})
		return
	}

	var media models.Media
	if err := h.Repo.DB.Preload("Planet").First(&media, mediaID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Медиа файл не найден"})
		return
	}

	ctx.JSON(http.StatusOK, media)
}

// UpdateMedia обновляет медиа файл
// @Summary        Обновить медиа файл
// @Description    Обновляет URL или тип медиа файла
// @Tags           media
// @Accept         json
// @Produce        json
// @Param          id      path    int          true    "ID медиа файла"
// @Param          media   body    models.Media true    "Обновленные данные"
// @Success        200     {object} models.Media
// @Failure        400     {object} map[string]string
// @Router         /api/media/{id} [put]
// @Security       BearerAuth
func (h *Handler) UpdateMedia(ctx *gin.Context) {
	mediaID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID медиа файла"})
		return
	}

	// Проверяем роль
	role := ctx.GetString("role")
	if role != "mission_control" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Только модератор может обновлять медиа файлы"})
		return
	}

	// Получаем существующий медиа файл
	var media models.Media
	if err := h.Repo.DB.First(&media, mediaID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Медиа файл не найден"})
		return
	}

	// Получаем обновленные данные
	var update struct {
		FileURL  string `json:"file_url"`
		FileType string `json:"file_type"`
	}

	if err := ctx.ShouldBindJSON(&update); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON"})
		return
	}

	// Проверяем тип файла
	if update.FileType != "" && update.FileType != "image" && update.FileType != "video" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Тип файла должен быть 'image' или 'video'"})
		return
	}

	// Обновляем поля
	if update.FileURL != "" {
		media.FileURL = update.FileURL
	}
	if update.FileType != "" {
		media.FileType = update.FileType
	}

	if err := h.Repo.DB.Save(&media).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления медиа файла"})
		return
	}

	ctx.JSON(http.StatusOK, media)
}
