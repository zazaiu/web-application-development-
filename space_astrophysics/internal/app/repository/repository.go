package repository

import "fmt"

type Planet struct {
	ID          int
	Name        string
	Perihelion  float64
	Aphelion    float64
	Description string
	ImageURL    string

	A      float64 // большая полуось (a.e.)
	E      float64 // эксцентриситет
	Period float64 // период (лет)
	T0     float64 // эпоха перигелия (JD)
}

type OrderedPlanet struct {
	Planet  Planet
	Comment string // комментарий пользователя
}

type Order struct {
	ID      int
	Planets []OrderedPlanet
	Date    string // дата заявки одна для всех
}

type Repository struct {
	Planets []Planet
	Orders  map[int]Order
}

func NewRepository() (*Repository, error) {
	planets := []Planet{
		{1, "Меркурий", 0.307, 0.467, "Ближайшая к Солнцу планета. Температура достигает 430°C днем и -180°C ночью. Не имеет атмосферы и спутников.",
			"http://172.19.0.2:9000/planets/mercury.png", 0.387, 0.206, 0.241, 2451547.5},
		{2, "Венера", 0.718, 0.728, "Вторая планета от Солнца. Имеет плотную атмосферу из углекислого газа с давлением в 92 раза больше земного. Температура поверхности около 465°C.",
			"http://172.19.0.2:9000/planets/venus.png", 0.723, 0.007, 0.615, 2451547.5},
		{3, "Земля", 0.983, 1.017, "Третья планета от Солнца. Единственная известная планета с жизнью. Имеет один естественный спутник - Луну.",
			"http://172.19.0.2:9000/planets/earth.png", 1.000, 0.017, 1.000, 2451547.5},
		{4, "Марс", 1.381, 1.666, "Четвертая планета, известная как 'Красная планета'. Имеет два спутника - Фобос и Деймос. Температура от -153°C до +20°C.",
			"http://172.19.0.2:9000/planets/mars.png", 1.524, 0.093, 1.881, 2451547.5},
		{5, "Юпитер", 4.950, 5.458, "Крупнейшая планета Солнечной системы, газовый гигант. Имеет 79 известных спутников, включая Ганимед - крупнейший спутник в Солнечной системе.",
			"http://172.19.0.2:9000/planets/jupiter.png", 5.203, 0.049, 11.86, 2451547.5},
		{6, "Сатурн", 9.041, 10.124, "Вторая по величине планета, известная своими кольцами. Имеет 82 спутника. Температура в верхних слоях атмосферы около -178°C.",
			"http://172.19.0.2:9000/planets/saturn.png", 9.537, 0.056, 29.46, 2451547.5},
		{7, "Уран", 18.286, 20.096, "Ледяной гигант с уникальным наклоном оси вращения (98°). Имеет 27 спутников и слабую систему колец. Температура около -224°C.",
			"http://172.19.0.2:9000/planets/uranus.png", 19.191, 0.046, 84.01, 2451547.5},
		{8, "Нептун", 29.81, 30.33, "Самая дальняя планета Солнечной системы, ледяной гигант. Имеет 14 спутников и самые сильные ветры в Солнечной системе - до 2100 км/ч.",
			"http://172.19.0.2:9000/planets/neptune.png", 30.07, 0.010, 164.8, 2451547.5},
	}

	worlds := make(map[int]Order)
	worlds[1] = Order{ID: 1, Planets: []OrderedPlanet{}, Date: ""}

	return &Repository{Planets: planets, Orders: worlds}, nil
}

func (r *Repository) GetAllPlanets() []Planet {
	return r.Planets
}

func (r *Repository) GetPlanetByID(id int) (Planet, error) {
	for _, p := range r.Planets {
		if p.ID == id {
			return p, nil
		}
	}
	return Planet{}, fmt.Errorf("планета с id=%d не найдена", id)
}

func (r *Repository) ViewMissionOrderByID(id int) (Order, error) {
	if world, ok := r.Orders[id]; ok {
		return world, nil
	}
	return Order{}, fmt.Errorf("заявка с id=%d не найдена", id)
}
