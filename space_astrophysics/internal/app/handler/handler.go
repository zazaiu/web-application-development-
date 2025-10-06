package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"space_astrophysics/internal/app/models"
	"space_astrophysics/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repo *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repo: r}
}

// ===== ВСПОМОГАТЕЛЬ: Получаем или создаём draft-заявку =====
func (h *Handler) getOrCreateDraftWorld(userID int) (*models.World, error) {
	draftWorld, err := h.Repo.GetDraftWorld(userID)
	if err != nil {
		return nil, err
	}
	if draftWorld == nil {
		newWorld := &models.World{
			CreatorID:   userID,
			WorldStatus: "draft",
		}
		if err := h.Repo.CreateWorld(newWorld); err != nil {
			return nil, err
		}
		draftWorld = newWorld
	}
	return draftWorld, nil
}

// ===== СПИСОК ПЛАНЕТ =====
func (h *Handler) ListPlanets(ctx *gin.Context) {
	q := ctx.Query("q")
	userID := 1 // пример пользователя

	// Получаем или создаём draft-заявку
	draftWorld, err := h.getOrCreateDraftWorld(userID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка работы с заявкой: %v", err)
		return
	}

	// Получаем все планеты
	planets, err := h.Repo.GetAllPlanets()
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка получения планет: %v", err)
		return
	}

	// Фильтр по имени
	var filtered []models.Planet
	if q == "" {
		filtered = planets
	} else {
		for _, p := range planets {
			if strings.Contains(strings.ToLower(p.Name), strings.ToLower(q)) {
				filtered = append(filtered, p)
			}
		}
	}

	// DTO для шаблона
	type PlanetDTO struct {
		ID         int
		Name       string
		ImageURL   string
		Perihelion float64
		Aphelion   float64
	}
	var planetsDTO []PlanetDTO
	for _, p := range filtered {
		image := "http://localhost:9000/planets/default-planet.png"
		if p.ImageURL != "" {
			image = "http://localhost:9000/planets/" + p.ImageURL
		}
		planetsDTO = append(planetsDTO, PlanetDTO{
			ID:         p.ID,
			Name:       p.Name,
			ImageURL:   image,
			Perihelion: p.A * (1 - p.E),
			Aphelion:   p.A * (1 + p.E),
		})
	}

	ctx.HTML(http.StatusOK, "service_list.html", gin.H{
		"Planets":   planetsDTO,
		"CartCount": len(draftWorld.Planets),
		"WorldID":   draftWorld.ID,
		"Q":         q,
	})
}

// ===== ДОБАВИТЬ ПЛАНЕТУ В ЧЕРНОВУЮ ЗАЯВКУ =====
func (h *Handler) AddPlanetToDraftWorld(ctx *gin.Context) {
	planetIDStr := ctx.PostForm("planet_id")
	planetID, err := strconv.Atoi(planetIDStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Неверный planet_id")
		return
	}

	// Проверяем, что планета существует
	_, err = h.Repo.GetPlanetByID(planetID)
	if err != nil {
		ctx.String(http.StatusNotFound, "Планета не найдена")
		return
	}

	userID := 1
	draftWorld, err := h.getOrCreateDraftWorld(userID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Добавляем планету
	err = h.Repo.AddPlanetToWorld(draftWorld.ID, planetID, 1, false)
	if err != nil {
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	ctx.Redirect(http.StatusFound, "/world/"+strconv.Itoa(draftWorld.ID))
}

// ===== VIEW WORLD =====
func (h *Handler) ViewWorld(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	world, err := h.Repo.GetWorldByID(id)
	if err != nil || world.WorldStatus == "deleted" {
		ctx.String(http.StatusNotFound, "Заявка не найдена или удалена")
		return
	}

	dateStr := ctx.Query("date")
	var now time.Time
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			now = parsed
		} else {
			now = time.Now()
		}
		world.Date = now.Format("2006-01-02")
	} else {
		now = time.Now()
		world.Date = now.Format("2006-01-02")
	}

	type Result struct {
		Planet models.Planet
		R_AU   float64
		NuDeg  float64
	}
	var results []Result
	for _, wp := range world.Planets {
		r, nu := calcOrbit(wp.Planet, now)
		results = append(results, Result{Planet: wp.Planet, R_AU: r, NuDeg: nu})
	}

	ctx.HTML(http.StatusOK, "order_detail.html", gin.H{
		"Order":   world,
		"Date":    world.Date,
		"Results": results,
	})
}

// ====== Расчёт орбиты =====
func julianDate(t time.Time) float64 {
	year, month, day := t.Date()
	if month <= 2 {
		year -= 1
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

func calcOrbit(p models.Planet, t time.Time) (float64, float64) {
	jd := julianDate(t)
	M := 2 * math.Pi * (jd - p.T0) / (p.Period * 365.25)
	M = math.Mod(M, 2*math.Pi)

	E := solveKepler(M, p.E)
	nu := 2 * math.Atan2(
		math.Sqrt(1+p.E)*math.Sin(E/2),
		math.Sqrt(1-p.E)*math.Cos(E/2),
	)
	r := p.A * (1 - p.E*math.Cos(E))
	return r, nu * 180 / math.Pi
}

// ===== ДЕТАЛИ ПЛАНЕТЫ =====
// ===== ДЕТАЛИ ПЛАНЕТЫ =====
func (h *Handler) ShowPlanetDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID планеты")
		return
	}

	planet, err := h.Repo.GetPlanetByID(id)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	// Для примера: userID=1
	userID := 1
	draftWorld, err := h.getOrCreateDraftWorld(userID)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Ошибка работы с заявкой: %v", err)
		return
	}

	now := time.Now()
	r, nu := calcOrbit(planet, now)

	// Используем прямой URL из MinIO
	image := planet.ImageURL
	if image == "" {
		image = "http://localhost:9000/planets/default-planet.png" // дефолтная картинка
	}

	ctx.HTML(http.StatusOK, "service_detail.html", gin.H{
		"Planet":    planet,
		"ImageURL":  image,
		"Date":      now.Format("2006-01-02"),
		"R_AU":      r,
		"NuDeg":     nu,
		"CartCount": len(draftWorld.Planets),
		"WorldID":   draftWorld.ID,
	})
}

// ===== ЛОГИЧЕСКОЕ УДАЛЕНИЕ ЗАЯВКИ =====
func (h *Handler) DeleteWorld(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID заявки")
		return
	}

	if err := h.Repo.DeleteWorldSQL(id); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Redirect(http.StatusFound, "/planets")
}

// ===== ОФОРМИТЬ ЗАЯВКУ (draft -> formed) =====
func (h *Handler) FormWorld(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.String(http.StatusBadRequest, "Некорректный ID заявки")
		return
	}

	if err := h.Repo.UpdateWorldStatus(id, "formed"); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Redirect(http.StatusFound, "/world/"+strconv.Itoa(id))
}
