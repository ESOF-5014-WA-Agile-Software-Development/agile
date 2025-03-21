package metamask

import (
	"errors"

	"gorm.io/gorm"

	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/models"
)

func Exists(c *api.Context) (interface{}, *api.Error) {
	address := c.GinCtx.Param("address")

	count := int64(0)
	err := c.Runtime.Mysql.Model(&models.MetaMask{}).Where("address = ?", address).Count(&count).Error
	if err != nil {
		return nil, api.InternalServerError()
	}
	if count > 0 {
		return map[string]bool{"exists": true}, nil
	}
	return map[string]bool{"exists": false}, nil
}

func Nonce(c *api.Context) (interface{}, *api.Error) {
	address := c.GinCtx.Param("address")

	var metamask models.MetaMask
	err := c.Runtime.Mysql.Where("address = ?", address).First(&metamask).Error
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
