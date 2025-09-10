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
func (h *Handler) GetServices(ctx *gin.Context) {
	q := ctx.Query("q")
	planets := h.Repository.GetAllPlanets()

	// простой фильтр по имени
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

	ctx.HTML(http.StatusOK, "services_list.html", gin.H{
		"Services":  filtered,
		"CartCount": len(h.Repository.Orders[1].Planets),
		"OrderID":   1,
		"Q":         q,
	})
}

// ===== 2. ДЕТАЛИ ПЛАНЕТЫ =====
func (h *Handler) GetService(ctx *gin.Context) {
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
		"Date":     now.Format("02.01.2006"),
		"R_AU":     r,
		"NuDeg":    nu,
	})
}

// ===== 3. ЗАЯВКА =====
func (h *Handler) GetOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	order, err := h.Repository.GetOrderByID(id)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	total := 0.0
	for _, p := range order.Planets {
		total += p.Planet.Perihelion
	}

	ctx.HTML(http.StatusOK, "order_detail.html", gin.H{
		"Order":         order,
		"Date":          time.Now().Format("02.01.2006"),
		"TotalDistance": total,
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

	order := h.Repository.Orders[1]
	found := false

	for i, op := range order.Planets {
		if op.Planet.ID == planet.ID {
			order.Planets[i].Count++
			found = true
			break
		}
	}

	if !found {
		order.Planets = append(order.Planets, repository.OrderedPlanet{Planet: planet, Count: 1})
	}

	h.Repository.Orders[1] = order
	ctx.Redirect(http.StatusFound, "/order/1")
}

// ===== 6. РАССЧИТАТЬ ЗАЯВКУ =====
func (h *Handler) CalcOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	dateStr := ctx.Query("date")
	var now time.Time
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			now = parsed
		} else {
			now = time.Now()
		}
	} else {
		now = time.Now()
	}

	order, err := h.Repository.GetOrderByID(id)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	totalAngle := 0.0
	count := 0
	for _, op := range order.Planets {
		r, nu := calcOrbit(op.Planet, now)
		_ = r
		totalAngle += nu * float64(op.Count)
		count += op.Count
	}

	avgAngle := 0.0
	if count > 0 {
		avgAngle = totalAngle / float64(count)
	}

	ctx.HTML(http.StatusOK, "order_detail.html", gin.H{
		"Order":      order,
		"Date":       now.Format("2006-01-02"),
		"AvgAngle":   avgAngle,
		"HasPlanets": len(order.Planets) > 0,
	})
}

// ==== 7. Очистить заявку ====
func (h *Handler) ClearOrder(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	if _, ok := h.Repository.Orders[id]; ok {
		h.Repository.Orders[id] = repository.Order{ID: id, Planets: []repository.OrderedPlanet{}}
	}

	ctx.Redirect(http.StatusFound, "/order/"+idStr)
}

//

//

// Юлианская дата
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

// решаем уравнение Кеплера
func solveKepler(M, e float64) float64 {
	E := M
	for i := 0; i < 15; i++ {
		E = E - (E-e*math.Sin(E)-M)/(1-e*math.Cos(E))
	}
	return E
}

// расчёт расстояния и угла
func calcOrbit(p repository.Planet, t time.Time) (float64, float64) {
	jd := julianDate(t)
	M := 2 * math.Pi * (jd - p.T0) / (p.Period * 365.25) // средняя аномалия
	M = math.Mod(M, 2*math.Pi)

	E := solveKepler(M, p.E)

	// истинная аномалия
	nu := 2 * math.Atan2(
		math.Sqrt(1+p.E)*math.Sin(E/2),
		math.Sqrt(1-p.E)*math.Cos(E/2),
	)

	// расстояние
	r := p.A * (1 - p.E*math.Cos(E))

	return r, nu * 180 / math.Pi
}
