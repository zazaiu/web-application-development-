package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"space_astrophysics/internal/app/models"
	"space_astrophysics/internal/app/repository"
	"space_astrophysics/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	Repo *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repo: r}
}

//=====================
//Сессии
//==========

// =========================================================
// 🔐 JWT + роли
// =========================================================

var jwtKey = []byte("super_secret_key")

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Middleware проверки токена и роли
func (h *Handler) AuthMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			c.Abort()
			return
		}

		// Проверка роли
		allowed := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// =========================================================
// 👤 USERS (регистрация / вход / профиль)
// =========================================================
// @Summary      Регистрация пользователя
// @Description  Создаёт нового пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "Пользователь"
// @Success      201   {object}  models.User
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/users/register [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var u models.User
	if err := ctx.ShouldBindJSON(&u); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if u.Role == "" {
		u.Role = "astronaut"
	}
	if err := h.Repo.CreateUser(&u); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, u)
}

func (h *Handler) gin(ctx *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.Repo.Authenticate(req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные"})
		return
	}

	// Генерация JWT
	tokenString, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании токена"})
		return
	}

	// Сохраняем сессию в Redis на 2 часа
	_ = utils.SetSession(tokenString, user.ID, user.Role, 2*time.Hour)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Успешный вход",
		"token":   tokenString,
		"role":    user.Role,
	})
}

// @Summary      Профиль пользователя
// @Description  Получение данных профиля авторизованного пользователя
// @Tags         users
// @Produce      json
// @Success      200  {object}  models.User
// @Failure      404  {object}  map[string]string
// @Router       /api/users/profile [get]
// @Security     ApiKeyAuth
func (h *Handler) GetUserProfile(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	user, err := h.Repo.GetUserByID(userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// =========================================================
// 🪐 PLANETS (услуги)
// =========================================================
// @Summary      Список планет
// @Description  Получение всех планет, опционально фильтруя по названию
// @Tags         planets
// @Produce      json
// @Param        q  query  string  false  "Поиск по имени"
// @Success      200  {array}   models.Planet
// @Failure      500  {object}  map[string]string
// @Router       /api/planets [get]
func (h *Handler) ListPlanets(ctx *gin.Context) {
	q := ctx.Query("q")
	planets, err := h.Repo.GetAllPlanets()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var filtered []models.Planet
	for _, p := range planets {
		if q == "" || strings.Contains(strings.ToLower(p.Name), strings.ToLower(q)) {
			filtered = append(filtered, p)
		}
	}
	ctx.JSON(http.StatusOK, filtered)
}

// @Summary      Детали планеты
// @Description  Получение данных конкретной планеты по ID
// @Tags         planets
// @Produce      json
// @Param        id  path  int  true  "ID планеты"
// @Success      200  {object}  models.Planet
// @Failure      404  {object}  map[string]string
// @Router       /api/planets/{id} [get]
func (h *Handler) ShowPlanetDetail(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	planet, err := h.Repo.GetPlanetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Планета не найдена"})
		return
	}
	ctx.JSON(http.StatusOK, planet)
}

// @Summary      Создание планеты
// @Description  Создаёт новую планету (только модератор)
// @Tags         planets
// @Accept       json
// @Produce      json
// @Param        planet  body  models.Planet  true  "Новая планета"
// @Success      201     {object}  models.Planet
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /api/planets [post]
// @Security     ApiKeyAuth
func (h *Handler) CreatePlanet(ctx *gin.Context) {
	var p models.Planet
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат"})
		return
	}
	p.Status = "active"
	if err := h.Repo.CreatePlanet(&p); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, p)
}

// @Summary      Обновление планеты
// @Description  Обновляет данные планеты по ID (только модератор)
// @Tags         planets
// @Accept       json
// @Produce      json
// @Param        id      path  int          true  "ID планеты"
// @Param        planet  body  models.Planet  true  "Данные для обновления"
// @Success      200     {object}  map[string]string
// @Failure      400     {object}  map[string]string
// @Failure      500     {object}  map[string]string
// @Router       /api/planets/{id} [put]
// @Security     ApiKeyAuth
func (h *Handler) UpdatePlanet(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var update models.Planet
	if err := ctx.ShouldBindJSON(&update); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.UpdatePlanet(id, update); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Планета обновлена"})
}

// @Summary      Удаление планеты
// @Description  Удаляет планету по ID (только модератор)
// @Tags         planets
// @Produce      json
// @Param        id  path  int  true  "ID планеты"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/planets/{id} [delete]
// @Security     ApiKeyAuth
func (h *Handler) DeletePlanet(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repo.DeletePlanet(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Планета удалена"})
}

// =========================================================
// 🌍 WORLDS (заявки)
// =========================================================

func (h *Handler) FormWorld(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repo.FormWorld(id); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка оформлена"})
}

// GET /api/worlds/:id
// @Summary Просмотр заявки
// @Description Получить подробную информацию о заявке по ID
// @Tags Worlds
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} models.World
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Security BearerAuth
// @Router /api/worlds/{id} [get]
func (h *Handler) ViewWorld(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	world, err := h.Repo.GetWorldByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Заявка не найдена"})
		return
	}
	ctx.JSON(http.StatusOK, world)
}

// PUT /api/worlds/:id
// @Summary Редактирование заявки
// @Description Обновить поля заявки (кроме ID, creator_id и статуса)
// @Tags Worlds
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param body body map[string]interface{} true "Поля для обновления"
// @Success 200 {object} map[string]string "Заявка обновлена"
// @Failure 400 {object} map[string]string "Некорректные данные"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Security BearerAuth
// @Router /api/worlds/{id} [put]
func (h *Handler) UpdateWorld(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	var update map[string]interface{}
	if err := ctx.ShouldBindJSON(&update); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	delete(update, "id")
	delete(update, "creator_id")
	delete(update, "world_status")
	if err := h.Repo.UpdateWorldFields(id, update); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка обновлена"})
}

// @Summary Список заявок
// @Description Получить список всех заявок. Для создателя — только свои, для модератора — все.
// @Tags Worlds
// @Accept json
// @Produce json
// @Param status query string false "Статус заявки (draft, formed, completed)"
// @Param from query string false "Дата начала фильтра YYYY-MM-DD"
// @Param to query string false "Дата конца фильтра YYYY-MM-DD"
// @Success 200 {array} models.World
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Security BearerAuth
// @Router /api/worlds [get]
func (h *Handler) ListWorldsFiltered(ctx *gin.Context) {
	role := ctx.GetString("role")
	userID := ctx.GetInt("user_id")

	status := ctx.Query("status")
	from := ctx.Query("from")
	to := ctx.Query("to")

	var worlds []models.World
	var err error

	if role == "mission_control" {
		worlds, err = h.Repo.GetWorldsFiltered(status, from, to)
	} else {
		worlds, err = h.Repo.GetWorldsByCreator(userID)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, worlds)
}

// DELETE /api/worlds/:id
// @Summary Удаление заявки
// @Description Устанавливает статус заявки 'deleted'
// @Tags Worlds
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string "Заявка удалена"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Security BearerAuth
// @Router /api/worlds/{id} [delete]
func (h *Handler) DeleteWorld(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repo.DeleteWorldSQL(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка удалена"})
}

// @Summary Завершение заявки
// @Description Завершает заявку, устанавливает total_cost и completed_at (только для mission_control)
// @Tags Worlds
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string "Заявка завершена"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 400 {object} map[string]string "Ошибка при завершении"
// @Security BearerAuth
// @Router /api/worlds/{id}/complete [put]
func (h *Handler) CompleteWorld(ctx *gin.Context) {
	role := ctx.GetString("role")
	if role != "mission_control" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещён"})
		return
	}
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := h.Repo.CompleteWorld(id, 2); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка завершена"})
}

// =========================================================
// ⚙️ M-M связи (WorldPlanet)
// =========================================================

// DELETE /api/worlds/:world_id/planet/:planet_id
/*func (h *Handler) DeleteWorldPlanet(ctx *gin.Context) {
	worldID, _ := strconv.Atoi(ctx.Param("world_id"))
	planetID, _ := strconv.Atoi(ctx.Param("planet_id"))
	if err := h.Repo.DeleteWorldPlanet(worldID, planetID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Планета удалена из заявки"})
}*/

// PUT /api/worlds/:world_id/planet/:planet_id
func (h *Handler) UpdateWorldPlanet(ctx *gin.Context) {
	worldID, _ := strconv.Atoi(ctx.Param("world_id"))
	planetID, _ := strconv.Atoi(ctx.Param("planet_id"))
	var payload struct {
		Quantity int  `json:"quantity"`
		IsMain   bool `json:"is_main"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON"})
		return
	}
	if err := h.Repo.UpdateWorldPlanet(worldID, planetID, payload.Quantity, payload.IsMain); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Связь обновлена"})
}

// DELETE /api/worlds/:world_id/planet/:planet_id
// @Summary Удаление планеты из заявки
// @Description Удаляет связь мир–планета
// @Tags Worlds
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param planet_id path int true "ID планеты"
// @Success 200 {object} map[string]string "Планета удалена из заявки"
// @Failure 400 {object} map[string]string "Некорректные данные"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Security BearerAuth
// @Router /api/worlds/{id}/planet/{planet_id} [delete]
func (h *Handler) DeleteWorldPlanet(ctx *gin.Context) {
	worldID, _ := strconv.Atoi(ctx.Param("world_id"))
	planetID, _ := strconv.Atoi(ctx.Param("planet_id"))
	if err := h.Repo.DeleteWorldPlanet(worldID, planetID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Планета удалена из заявки"})
}

// @Summary      Обновление профиля
// @Description  Обновляет данные авторизованного пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "Обновлённые данные"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/users/profile [put]
// @Security     ApiKeyAuth
func (h *Handler) UpdateUserProfile(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	var update models.User
	if err := ctx.ShouldBindJSON(&update); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON"})
		return
	}
	if err := h.Repo.UpdateUser(userID, update); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Профиль обновлён"})
}

// @Summary      Вход пользователя
// @Description  Аутентификация и выдача JWT
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        credentials  body      map[string]string  true  "Логин и пароль"
// @Success      200          {object}  map[string]string
// @Failure      400          {object}  map[string]string
// @Failure      401          {object}  map[string]string
// @Failure      500          {object}  map[string]string
// @Router       /api/users/login [post]
func (h *Handler) Login(ctx *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.Repo.Authenticate(req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные"})
		return
	}

	tokenString, err := utils.GenerateJWT(user.ID, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании токена"})
		return
	}

	if err := utils.SetSession(tokenString, user.ID, user.Role, 2*time.Hour); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении сессии"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Успешный вход",
		"token":   tokenString,
		"role":    user.Role,
	})
}

// @Summary      Выход пользователя
// @Description  Удаляет сессию из Redis
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Router       /api/users/logout [post]
// @Security     ApiKeyAuth
func (h *Handler) Logout(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Требуется токен"})
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")
	_ = utils.DeleteSession(token)
	ctx.JSON(http.StatusOK, gin.H{"message": "Вы вышли"})
}

// =========================================================
// 🛒 КОРЗИНА / WORLDS
// =========================================================

func (h *Handler) GetCartIcon(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	count, err := h.Repo.GetUserCartCount(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"cart_count": count})
}

// @Summary Добавление планеты в заявку
// @Description Добавляет планету в заявку (черновую или существующую)
// @Tags Worlds
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param planet_id path int true "ID планеты"
// @Param body body struct{Angle float64; Distance float64; Comment string} true "Данные планеты"
// @Success 200 {object} map[string]interface{} "Планета добавлена в заявку"
// @Failure 400 {object} map[string]string "Некорректные данные"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Security BearerAuth
// @Router /api/worlds/{id}/planet/{planet_id} [post]
func (h *Handler) AddPlanetToWorld(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")
	planetID, _ := strconv.Atoi(ctx.Param("id"))

	var req struct {
		Angle    float64 `json:"angle"`
		Distance float64 `json:"distance"`
		Comment  string  `json:"comment"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный JSON"})
		return
	}

	world, err := h.Repo.GetOrCreateDraftWorld(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repo.AddPlanetToWorld(world.ID, planetID, req.Angle, req.Distance, req.Comment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "Планета добавлена в заявку",
		"world_id": world.ID,
	})
}

// =========================================================
// 🖼️ PLANET IMAGE
// =========================================================

func (h *Handler) UploadPlanetImage(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Файл не загружен"})
		return
	}
	path := "static/uploads/" + file.Filename
	if err := ctx.SaveUploadedFile(file, path); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.Repo.UpdatePlanetImage(id, "/"+path); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Изображение обновлено", "path": "/" + path})
}

// =========================================================
// 🔗 M-M GET /world-planets/:world_id/:planet_id
// =========================================================

func (h *Handler) GetWorldPlanet(ctx *gin.Context) {
	worldID, _ := strconv.Atoi(ctx.Param("world_id"))
	planetID, _ := strconv.Atoi(ctx.Param("planet_id"))
	wp, err := h.Repo.GetWorldPlanet(worldID, planetID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Связь не найдена"})
		return
	}
	ctx.JSON(http.StatusOK, wp)
}

// =========================================================
// 🚀 Утилиты расчёта орбит
// =========================================================

func julianDate(t time.Time) float64 {
	year, month, day := t.Date()
	if month <= 2 {
		year--
		month += 12
	}
	A := year / 100
	B := 2 - A + A/4
	return float64(int(365.25*float64(year+4716))) +
		float64(int(30.6001*float64(month+1))) +
		float64(day) + float64(B) - 1524.5
}

func solveKepler(M, e float64) float64 {
	E := M
	for i := 0; i < 15; i++ {
		E = E - (E-e*math.Sin(E)-M)/(1-e*math.Cos(E))
	}
	return E
}
