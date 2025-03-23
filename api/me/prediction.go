package me

import (
	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/models"
)

var input struct {
	Storage     float64 `json:"storage"`
	Capacity    float64 `json:"capacity"`
	Generation  float64 `json:"generation"`
	Consumption float64 `json:"consumption"`
	Saleable    float64 `json:"saleable"`
}

func UpdatePrediction(c *api.Context) (interface{}, *api.Error) {
	id := c.GinCtx.Param("id")

	if err := c.GinCtx.ShouldBindJSON(&input); err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	var prediction models.Prediction
	if err := c.Runtime.Mysql.First(&prediction, id).Error; err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	prediction.Capacity = input.Capacity
	prediction.Storage = input.Storage
	prediction.Generation = input.Generation
	prediction.Consumption = input.Consumption
	prediction.Saleable = input.Saleable

	if err := c.Runtime.Mysql.Save(&prediction).Error; err != nil {
		return nil, api.InternalServerError()
	}

	return prediction, nil
}

func GetPrediction(c *api.Context) (interface{}, *api.Error) {
	id := c.GinCtx.Param("id")

	var prediction models.Prediction
	err := c.Runtime.Mysql.First(&prediction, id).Error

	if err != nil {
		return nil, api.InternalServerError()
	}

	return prediction, nil
}
