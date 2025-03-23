package models

type Prediction struct {
	Model

	Capacity    float64 `json:"capacity" gorm:"type:decimal(20,2);not null"`
	Storage     float64 `json:"storage" gorm:"type:decimal(20,2);not null"`
	Generation  float64 `json:"generation" gorm:"type:decimal(20,2);not null"`
	Consumption float64 `json:"consumption" gorm:"type:decimal(20,2);not null"`
	Saleable    float64 `json:"saleable" gorm:"type:decimal(20,2);not null"`
}
