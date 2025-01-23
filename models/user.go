package models

type Role string

const (
	RoleDefault Role = "default"
	RoleAdmin   Role = "admin"
)

type User struct {
	Model

	Name      string     `json:"name" gorm:"type:varchar(64);unique;not null"`
	Email     string     `json:"email" gorm:"type:varchar(320);unique"`
	Password  string     `json:"password" gorm:"type:varchar(64)"`
	Role      string     `json:"role" gorm:"type:varchar(64);not null"`
	MetaMasks []MetaMask `json:"meta_masks" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type MetaMask struct {
	Model
	UserID uint

	Address string `json:"address" gorm:"type:varchar(64);unique;not null"`
	Nonce   string `json:"nonce" gorm:"type:varchar(8);not null"`
}
