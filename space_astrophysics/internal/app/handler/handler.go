package handler

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"space_astrophysics/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// ===== 1. СПИСОК ПЛАНЕТ =====
func (h *Handler) ListPlanets(ctx *gin.Context) {
	q := ctx.Query("q")
	planets := h.Repository.GetAllPlanets()

	// фильтр по имени
	var filtered []repository.Planet
	if q == "" {
		filtered = planets
	} else {
		for _, p := range planets {
			if strings.Contains(strings.ToLower(p.Name), strings.ToLower(q)) {
				filtered = append(filtered, p)
			}
		}
	}

	ctx.HTML(http.StatusOK, "service_list.html", gin.H{
		"Planets":   filtered, // ✅ заменили Services → Planets
		"CartCount": len(h.Repository.Orders[1].Planets),
		"OrderID":   1,
		"Q":         q,
	})
}

// ===== 2. ДЕТАЛИ ПЛАНЕТЫ =====
func (h *Handler) ShowPlanetDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	planet, err := h.Repository.GetPlanetByID(id)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	now := time.Now()
	r, nu := calcOrbit(planet, now)

	ctx.HTML(http.StatusOK, "service_detail.html", gin.H{
		"Planet":   planet,
		"ImageURL": planet.ImageURL,
		"Date":     now.Format("2006-01-02"),
		"R_AU":     r,
		"NuDeg":    nu,
	})
}

// ===== 3. ЗАЯВКА (world) =====
func (h *Handler) ViewMissionOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	world, err := h.Repository.ViewMissionOrderByID(id)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	// читаем дату из query
	dateStr := ctx.Query("date")
	var now time.Time
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			now = parsed
		} else {
			now = time.Now()
		}
		world.Date = dateStr
	} else {
		now = time.Now()
		world.Date = now.Format("2006-01-02")
	}

	// считаем угол и расстояние для каждой планеты
	type Result struct {
		Planet repository.Planet
		R_AU   float64
		NuDeg  float64
	}
	var results []Result
	for _, op := range world.Planets {
		r, nu := calcOrbit(op.Planet, now)
		results = append(results, Result{Planet: op.Planet, R_AU: r, NuDeg: nu})
	}

	ctx.HTML(http.StatusOK, "order_detail.html", gin.H{
		"Order":   world,
		"Date":    world.Date,
		"Results": results,
	})
}

// ===== 4. ДОБАВИТЬ В ЗАЯВКУ =====
func (h *Handler) AddToOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	planet, err := h.Repository.GetPlanetByID(id)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	world := h.Repository.Orders[1]
	found := false
	for _, op := range world.Planets {
		if op.Planet.ID == planet.ID {
			found = true
			break
		}
	}

	if !found {
		world.Planets = append(world.Planets, repository.OrderedPlanet{Planet: planet, Comment: ""})
	}

	h.Repository.Orders[1] = world
	ctx.Redirect(http.StatusFound, "/world/1")
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

func calcOrbit(p repository.Planet, t time.Time) (float64, float64) {
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
