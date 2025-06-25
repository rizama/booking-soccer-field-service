package middlewares

import (
	"crypto/sha256"
	"encoding/hex"
	"field-service/clients"
	"field-service/common/response"
	"field-service/config"
	"field-service/constants"
	errConstant "field-service/constants/error"
	"fmt"
	"net/http"
	"slices"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func HandlePanic() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("Recovered from panic: %v", err)
				ctx.JSON(http.StatusInternalServerError, response.Response{
					Status:  constants.Error,
					Message: errConstant.ErrInternalServerError.Error(),
				})
				ctx.Abort()
			}
		}()
		ctx.Next()
	}
}

func RateLimit(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		err := tollbooth.LimitByRequest(lmt, ctx.Writer, ctx.Request)
		if err != nil {
			ctx.JSON(http.StatusTooManyRequests, response.Response{
				Status:  constants.Error,
				Message: errConstant.ErrToManyRequest.Error(),
			})
			ctx.Abort()
		}
		ctx.Next()
	}
}

func responseUnauthorized(ctx *gin.Context, message string) {
	ctx.JSON(http.StatusUnauthorized, response.Response{
		Status:  constants.Error,
		Message: message,
	})
	ctx.Abort()
}

func contains(roles []string, role string) bool {
	return slices.Contains(roles, role)
}

func CheckRole(roles []string, client clients.IClientRegistry) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := client.UserSvc().GetUserByToken(ctx.Request.Context())
		if err != nil {
			responseUnauthorized(ctx, errConstant.ErrUnauthorized.Error())
			return
		}

		if !contains(roles, user.Role) {
			responseUnauthorized(ctx, errConstant.ErrForbidden.Error())
			return
		}

		ctx.Next()
	}
}

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		token := c.GetHeader(constants.Authorization)
		if token == "" {
			responseUnauthorized(c, errConstant.ErrUnauthorized.Error())
			return
		}

		err = validateApiKey(c)
		if err != nil {
			responseUnauthorized(c, err.Error())
			return
		}

		c.Next()
	}
}

func AuthenticateWithoutToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := validateApiKey(c)
		if err != nil {
			responseUnauthorized(c, err.Error())
			return
		}

		c.Next()
	}
}

func validateApiKey(ctx *gin.Context) error {
	apiKey := ctx.GetHeader(constants.XApiKey)
	requestAt := ctx.GetHeader(constants.XRequestAt)
	serviceName := ctx.GetHeader(constants.XServiceName)
	signatureKey := config.Config.SignatureKey

	validateKey := fmt.Sprintf("%s:%s:%s", serviceName, signatureKey, requestAt)
	hash := sha256.New()
	hash.Write([]byte(validateKey))
	resultHash := hex.EncodeToString(hash.Sum(nil))

	if apiKey != resultHash {
		return errConstant.ErrUnauthorized
	}
	return nil
}
