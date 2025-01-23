package emails

import (
	"github.com/gin-gonic/gin"

	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/models"
	"github.com/mylakehead/agile/runtime"
)

func Exists(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	email := c.Param("email")

	count := int64(0)
	err := rt.Mysql.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return nil, api.InternalServerError()
	}
	if count > 0 {
		return map[string]bool{"exists": true}, nil
	}
	return map[string]bool{"exists": false}, nil
}
