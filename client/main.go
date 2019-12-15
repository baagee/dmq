package main

import (
	"dmq/message"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

//响应结构体
type ResponseData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func main() {
	http.HandleFunc("/send", send)
	http.ListenAndServe(":8989", nil)
}

func send(writer http.ResponseWriter, request *http.Request) {
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
		resp.Message = "Error: " + err.Error()
	} else {
		ts, _ := strconv.Atoi(params["timestamp"].(string))
		msg := message.Message{
			Id:        getId(),
			Project:   params["project"].(string),
			Cmd:       params["cmd"].(string),
			Timestamp: uint64(time.Unix(int64(ts), 0).Unix()),
			Params:    params["params"].(string),
			List:      params["list"].(string),
		}
		log.Printf("%v\n", msg)
		msg.Save()
	}
	bb, _ := json.Marshal(&resp)
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Write(bb)
}

func getId() uint64 {
	return uint64(time.Now().Unix()*1000 + rand.Int63n(8999) + 1000)
}
