package users

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"

	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/lib"
	"github.com/mylakehead/agile/models"
	"github.com/mylakehead/agile/runtime"
)

const (
	signInTypeEmail     string = "email"
	signInTypeMetamask  string = "metamask"
	signInTypePhone     string = "phone"
	signInTypeTwitter   string = "twitter"
	signInTypeFacebook  string = "facebook"
	signInTypeInstagram string = "instagram"
)

type signInByMetaMaskRequest struct {
	MetaMask string `json:"metamask" binding:"required"`
	Sign     string `json:"sign" binding:"required"`
}

func SignIn(c *api.Context) (interface{}, *api.Error) {
	t := c.GinCtx.Param("type")

	switch t {
	case signInTypeMetamask:
		return signInByMetaMask(c.Runtime, c.GinCtx)
	case signInTypeEmail:
		fallthrough
	case signInTypePhone:
		fallthrough
	case signInTypeTwitter:
		fallthrough
	case signInTypeFacebook:
		fallthrough
	case signInTypeInstagram:
		fallthrough
	default:
		return nil, api.InvalidArgument(nil, "invalid sign in type")
	}
}

func signInByMetaMask(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	req := signInByMetaMaskRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	var metamask models.MetaMask
	err := rt.Mysql.Where("address = ?", req.MetaMask).First(&metamask).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, api.InvalidArgument(nil, "unsigned up address")
		}
		return nil, api.InternalServerError()
	}

	// check sign
	data := fmt.Sprintf("Agile is signing the nonce: %s for login.", metamask.Nonce)
	sig, err := hexutil.Decode(req.Sign)
	if err != nil {
		return nil, api.InvalidArgument(nil, "sign check error")
	}
	addr, err := ecRecover(hexutil.Bytes(data), sig)
	if err != nil {
		return nil, api.InvalidArgument(nil, "sign check error")
	}
	if !hasMatchingAddress(addr.String(), req.MetaMask) {
		return nil, api.InvalidArgument(nil, "address check error")
	}
	// refresh nonce
	nonce, err := lib.GenerateCaptcha(8)
	if err != nil {
		return nil, api.InternalServerError()
	}
	metamask.Nonce = nonce
	rt.Mysql.Save(metamask)

	// get user
	var user models.User
	err = rt.Mysql.First(&user, metamask.UserID).Error
	if err != nil {
		return nil, api.InternalServerError()
	}

	// get metaMasks
	var metaMasks []models.MetaMask
	err = rt.Mysql.Where("user_id = ?", user.ID).Find(&metaMasks).Error
	if err != nil {
		return nil, api.InternalServerError()
	}
	ms := make([]string, 0)
	for _, m := range metaMasks {
		ms = append(ms, m.Address)
	}

	// jwt
	expire := time.Second * time.Duration(rt.Config.Jwt.Expire)
	claims := lib.JWTClaims{
		UserId:    user.ID,
		UserName:  user.Name,
		UserEmail: user.Email,
		UserRole:  user.Role,
		MetaMasks: ms,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expire).Unix(),
			Id:        strconv.Itoa(int(user.ID)),
			IssuedAt:  0,
			Issuer:    "agile.lakehead",
			Subject:   "user",
			Audience:  "",
			NotBefore: 0,
		},
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := at.SignedString([]byte(rt.Config.Jwt.Key))
	if err != nil {
		return nil, api.InternalServerError("create token error")
	}

	// TODO refresh token
	return map[string]interface{}{
		"user": map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
		"token": token,
	}, nil
}
