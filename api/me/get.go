package me

import (
	"time"

	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/models"
)

func Ongoing(c *api.Context) (interface{}, *api.Error) {
	now := time.Now().Unix()
	timeLimit := now - 1800

	var transactions []models.Purchased

	err := c.Runtime.Mysql.Where("timestamp >= ?", timeLimit).
		Order("timestamp DESC").
		Find(&transactions).Error

	if err != nil {
		return nil, api.InternalServerError()
	}

	return transactions, nil
}
