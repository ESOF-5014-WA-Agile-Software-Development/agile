package me

import (
	"time"

	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/models"
)

func Ongoing(c *api.Context) (interface{}, *api.Error) {
	now := time.Now().Unix()
	// timeLimit := now - 180000
	timeLimit := now - 1800

	var transactions []models.Purchased

	if len(c.MetaMasks) <= 0 {
		return nil, api.InvalidArgument(nil, "please bind your MetaMask wallet")
	}

	err := c.Runtime.Mysql.
		Where("timestamp >= ? AND (seller = ? OR buyer = ?)", timeLimit, c.MetaMasks[0], c.MetaMasks[0]).
		Order("timestamp DESC").
		Find(&transactions).Error

	if err != nil {
		return nil, api.InternalServerError()
	}

	return transactions, nil
}
