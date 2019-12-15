package handle

import (
	"dmq/message"
	"dmq/util"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func Single(writer http.ResponseWriter, request *http.Request) {
	jsonBytes, _ := ioutil.ReadAll(request.Body)
	params := make(map[string]interface{})
	err := json.Unmarshal(jsonBytes, &params)

	resp := ResponseData{
		Code:    0,
		Message: "success",
		Data:    nil,
	}

	if err != nil {
		// 失败
		log.Println("json decode 失败: " + err.Error())
		resp.Code = 10010
		resp.Message = "json decode error: " + err.Error()
	} else {
		mid, err := save(params)
		if err != nil {
			log.Println("保存失败: " + err.Error())
			resp.Code = 10086
			resp.Message = "Error: " + err.Error()
		} else {
			resp.Data = mid
			log.Println("保存成功")
		}
	}
	bb, _ := json.Marshal(&resp)
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(bb)
}

func save(params map[string]interface{}) (uint64, error) {
	ts, _ := strconv.Atoi(params["timestamp"].(string))
	msg := message.Message{
		Id:        util.GetNumberId(),
		Project:   params["project"].(string),
		Cmd:       params["cmd"].(string),
		Timestamp: uint64(time.Unix(int64(ts), 0).Unix()),
		Params:    params["params"].(string),
		Bucket:    params["bucket"].(string),
	}
	return msg.Id, msg.Save()
}
