package users

import (
	"context"
	"errors"
	"fmt"
	"github.com/mylakehead/agile/lib"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/sha3"
	"gorm.io/gorm"

	"github.com/mylakehead/agile/api"
	"github.com/mylakehead/agile/models"
	"github.com/mylakehead/agile/runtime"
)

const (
	signupTypeEmail     string = "email"
	signupTypeMetamask  string = "metamask"
	signupTypePhone     string = "phone"
	signupTypeTwitter   string = "twitter"
	signupTypeFacebook  string = "facebook"
	signupTypeInstagram string = "instagram"
)

type signupByMetaMaskRequest struct {
	MetaMask string `json:"metamask" binding:"required"`
	Sign     string `json:"sign" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Captcha  string `json:"captcha" binding:"required"`
}

func SignUp(c *api.Context) (interface{}, *api.Error) {
	t := c.GinCtx.Param("type")

	switch t {
	case signupTypeMetamask:
		return signupByMetaMask(c.Runtime, c.GinCtx)
	case signupTypeEmail:
		fallthrough
	case signupTypePhone:
		fallthrough
	case signupTypeTwitter:
		fallthrough
	case signupTypeFacebook:
		fallthrough
	case signupTypeInstagram:
		fallthrough
	default:
		return nil, api.InvalidArgument(nil, "invalid sign up type")
	}
}

func hasMatchingAddress(knownAddress string, recoveredAddress string) bool {
	return strings.ToLower(knownAddress) == strings.ToLower(recoveredAddress)
}

func TextHash(data []byte) []byte {
	hash, _ := TextAndHash(data)
	return hash
}

func TextAndHash(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), string(data))
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(msg))
	return h.Sum(nil), msg
}

func ecRecover(data, sig hexutil.Bytes) (common.Address, error) {
	if len(sig) != crypto.SignatureLength {
		return common.Address{}, fmt.Errorf("signature must be %d bytes long", crypto.SignatureLength)
	}
	if sig[crypto.RecoveryIDOffset] != 27 && sig[crypto.RecoveryIDOffset] != 28 {
		return common.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1

	rpk, err := crypto.SigToPub(TextHash(data), sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*rpk), nil
}

func signupByMetaMask(rt *runtime.Runtime, c *gin.Context) (interface{}, *api.Error) {
	// result: activated user
	//         with metamask address
	//         with email
	//         without phone
	req := signupByMetaMaskRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}

	// check sign
	data := "Agile uses this cryptographic signature to verify that you are the owner of this address."
	sig, err := hexutil.Decode(req.Sign)
	if err != nil {
		return nil, api.InvalidArgument(nil, "sign check error")
	}
	addr, err := ecRecover(hexutil.Bytes(data), sig)
	if err != nil {
		return nil, api.InvalidArgument(nil, err.Error())
	}
	if !hasMatchingAddress(addr.String(), req.MetaMask) {
		return nil, api.InvalidArgument(nil, "address check error")
	}

	// metamask address exists?
	count := int64(0)
	err = rt.Mysql.Model(&models.MetaMask{}).Where(
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

	// check captcha
	key := fmt.Sprintf("%s/%s/%s", req.Name, req.Email, req.MetaMask)
	captcha, err := rt.Redis.Cli.Get(context.TODO(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, api.InvalidArgument(nil, "invalid captcha")
		} else {
			return nil, api.InternalServerError()
		}
	}
	if captcha != req.Captcha {
		return nil, api.InvalidArgument(nil, "invalid captcha")
	}

	// insert records
	nonce, err := lib.GenerateCaptcha(8)
	if err != nil {
		return nil, api.InternalServerError()
	}
	err = rt.Mysql.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&models.User{
			Name:  req.Name,
			Email: req.Email,
			Role:  string(models.RoleDefault),
			MetaMasks: []models.MetaMask{
				{
					Address: req.MetaMask,
					Nonce:   nonce,
				},
			},
		}).Error; err != nil {
			// return any error will roll back
			return err
		}
		// return nil will commit the whole transaction
		return nil
	})
	if err != nil {
		return nil, api.InternalServerError()
	}

	// clear redis
	err = rt.Redis.Cli.Del(context.TODO(), key).Err()
	if err != nil {
		println(err.Error())
	}

	return nil, nil
}
