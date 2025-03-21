package users

import (
	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/models"
)

func Exists(c *api.Context) (interface{}, *api.Error) {
	name := c.GinCtx.Param("name")

	count := int64(0)
	err := c.Runtime.Mysql.Model(&models.User{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return nil, api.InternalServerError()
	}
	if count > 0 {
		return map[string]bool{"exists": true}, nil
	}

	return map[string]bool{"exists": false}, nil
}
