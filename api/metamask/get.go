package metamask

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/models"
	"github.com/mylakehead/agile/runtime"
)

func Exists(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	address := c.Param("address")

	count := int64(0)
	err := rt.Mysql.Model(&models.MetaMask{}).Where("address = ?", address).Count(&count).Error
	if err != nil {
		return nil, api.InternalServerError()
	}
	if count > 0 {
		return map[string]bool{"exists": true}, nil
	}
	return map[string]bool{"exists": false}, nil
}

func Nonce(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	address := c.Param("address")

	var metamask models.MetaMask
	err := rt.Mysql.Where("address = ?", address).First(&metamask).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.NotFoundError()
		}

		return nil, api.InternalServerError()
	}

	return map[string]string{
		"nonce": metamask.Nonce,
	}, nil
}
