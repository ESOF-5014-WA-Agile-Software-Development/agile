package emails

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/lib"
	"github.com/mylakehead/agile/models"
	"github.com/mylakehead/agile/runtime"
)

const (
	verifyTypeEmail string = "email"
	verifyTypePhone string = "phone"
)

type VerifyEmailRequest struct {
	MetaMask string `json:"metamask" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

func Verify(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	t := c.Param("type")

	switch t {
	case verifyTypeEmail:
		return verifyEmail(rt, c)
	case verifyTypePhone:
		fallthrough
	default:
		return nil, api.InvalidArgument(nil, "invalid verify type")
	}
}

func verifyEmail(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	req := VerifyEmailRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	// metamask address exists?
	count := int64(0)
	err := rt.Mysql.Model(&models.MetaMask{}).Where(
		"address = ?", req.MetaMask).Count(&count).Error
	if err != nil {
		return nil, api.InternalServerError()
	}
	if count > 0 {
		return nil, api.InvalidArgument(nil, "metamask address exists")
	}

	// user
	count = int64(0)
	err = rt.Mysql.Model(&models.User{}).Where(
		"name = ?", req.Name).Count(&count).Error
	if err != nil {
		return nil, api.InternalServerError()
	}
	if count > 0 {
		return nil, api.InvalidArgument(nil, "user name exists")
	}

	// email
	count = int64(0)
	err = rt.Mysql.Model(&models.User{}).Where(
		"email = ?", req.Email).Count(&count).Error
	if err != nil {
		return nil, api.InternalServerError()
	}
	if count > 0 {
		return nil, api.InvalidArgument(nil, "user email exists")
	}

	// set redis key
	key := fmt.Sprintf("%s/%s/%s", req.Name, req.Email, req.MetaMask)
	captcha, err := lib.GenerateCaptcha(6)
	if err != nil {
		return nil, api.InternalServerError("encode captcha error")
	}

	err = rt.Redis.Cli.Set(context.TODO(), key, captcha, 2*time.Hour).Err()
	if err != nil {
		return nil, api.InternalServerError("redis error")
	}

	err = rt.Email.Send(captcha, req.Email)
	if err != nil {
		return nil, api.InternalServerError("send email error")
	}

	return nil, nil
}
