package response

import (
	"field-service/constants"
	"net/http"

	errConsttant "field-service/constants/error"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  string  `json:"status"`
	Message any     `json:"message"`
	Data    any     `json:"data"`
	Token   *string `json:"token,omitempty"`
}

type ParamHttpResp struct {
	Code    int
	Err     error
	Message *string
	Gin     *gin.Context
	Data    any
	Token   *string
}

func HttpRresponse(param ParamHttpResp) {
	if param.Err == nil {
		param.Gin.JSON(param.Code, Response{
			Status:  constants.Success,
			Message: http.StatusText(param.Code),
			Data:    param.Data,
			Token:   param.Token,
		})
		return
	}

	// if error
	message := errConsttant.ErrInternalServerError.Error()
	if param.Message != nil {
		message = *param.Message
	} else if param.Err != nil {
		if errConsttant.ErrMapping(param.Err) {
			message = param.Err.Error()
		}
	}

	param.Gin.JSON(param.Code, Response{
		Status:  constants.Failed,
		Message: message,
		Data:    param.Data,
	})

	return
}
