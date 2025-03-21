package models

type Purchased struct {
	Model

	BlockID uint64 `json:"block_id" gorm:"unique;not null"`
	OfferID uint64 `json:"offer_id" gorm:"index;not null"`

	Seller string  `json:"seller" gorm:"type:varchar(64);not null"`
	Buyer  string  `json:"buyer" gorm:"type:varchar(64);not null"`
	Amount float64 `json:"amount" gorm:"type:decimal(20,2);not null"`
}

func (Purchased) TableName() string {
	return "purchased"
}
