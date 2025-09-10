// data.go
package main

type Service struct {
	ID          int
	Name        string
	Perihelion  float64 // Ближайшая точка орбиты
	Aphelion    float64 // Самая дальняя точка орбиты
	Description string
	ImageName   string
}

type OrderItem struct {
	OrderID     int
	ServiceID   int
	Nights      int    // Количество расчетов
	CurrentDate string // Дата расчета
}

type Order struct {
	ID         int
	TotalPrice float64
}

var Services = map[int]Service{
	1: {ID: 1, Name: "Меркурий", Perihelion: 0.307, Aphelion: 0.467,
		Description: "Меркурий – самая близкая к Солнцу планета...", ImageName: "1.jpg"},
	2: {ID: 2, Name: "Венера", Perihelion: 0.718, Aphelion: 0.728,
		Description: "Описание Венеры...", ImageName: "2.jpg"},
}

var Orders = map[int]Order{
	101: {ID: 101, TotalPrice: 0},
}

var OrderItems = []OrderItem{
	{OrderID: 101, ServiceID: 1, Nights: 1, CurrentDate: "01.01.2025"},
}
