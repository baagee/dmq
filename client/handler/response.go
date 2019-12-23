package handler

import (
	"encoding/json"
	"github.com/baagee/dmq/common"
	"net/http"
)

//响应结构体
type responseBody struct {
	Code    uint        `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

//输出错误信息
func responseWithError(writer http.ResponseWriter, err error) {
	resp := responseBody{
		Code:    0,
		Message: err.Error(),
		Data:    nil,
	}
	common.RecordError(err)
	switch e := err.(type) {
	case common.Notice:
		//自定义的Error类型
		resp.Code = uint(e.Code())
		responseWithJson(writer, resp)
	default:
		resp.Code = common.ErrorCodeDefault
		responseWithJson(writer, resp)
	}
}

//输出json
func responseWithJson(writer http.ResponseWriter, respBody responseBody) {
	resp, _ := json.Marshal(respBody)
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, err := writer.Write(resp)
	if err != nil {
		common.RecordError(err)
	}
}
