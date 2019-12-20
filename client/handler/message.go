package handler

import (
	"encoding/json"
	"errors"
	"github.com/baagee/dmq/common"
	"log"
	"net/http"
	"time"
)

//当个请求的结构体
type singleRequest struct {
	Cmd       string `json:"cmd" validate:"required"`     // 命令点
	Timestamp uint64 `json:"timestamp"`                   // 执行时间
	Params    string `json:"params"`                      // 命令参数
	Project   string `json:"project" validate:"required"` // 项目
	Bucket    string `json:"bucket" validate:"required"`  // 消息桶
}

//响应结构体
type responseBody struct {
	Code    uint        `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

//批量请求结构体
type batchRequest []singleRequest

// 单个请求
func SingleMessage(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var single singleRequest
	if err := json.NewDecoder(request.Body).Decode(&single); err != nil {
		responseWithError(writer, common.ThrowNotice(common.ErrorCodeParseParamsFailed, errors.New("不是合法的json")))
		return
	}

	var singleList batchRequest
	singleList = append(singleList, single)
	// 验证参数
	if err := checkParams(singleList, common.GetClientIP(request)); err != nil {
		responseWithError(writer, err)
		return
	}

	//保存
	var respBody responseBody
	respBody.Data = save(singleList)[0]
	responseWithJson(writer, respBody)
}

// 批量请求
func BatchMessage(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var singleList batchRequest
	if err := json.NewDecoder(request.Body).Decode(&singleList); err != nil {
		responseWithError(writer, common.ThrowNotice(common.ErrorCodeParseParamsFailed, errors.New("不是合法的json")))
		return
	}
	// 验证参数
	if err := checkParams(singleList, common.GetClientIP(request)); err != nil {
		responseWithError(writer, err)
		return
	}

	//保存
	var respBody responseBody
	respBody.Data = save(singleList)
	responseWithJson(writer, respBody)
}

//保存命令
func save(singleList batchRequest) []bool {
	length := len(singleList)
	// 切片需要make
	ret := make([]bool, length)
	ids := common.GenerateIds(int64(length))
	for i, single := range singleList {
		if single.Timestamp == 0 {
			// 时间为0 表示当前时间 立即执行
			single.Timestamp = uint64(time.Now().Unix())
		}
		msg := common.Message{
			Id:         ids[i],
			Cmd:        single.Cmd,
			Timestamp:  single.Timestamp,
			Params:     single.Params,
			Project:    single.Project,
			Bucket:     single.Bucket,
			CreateTime: uint64(time.Now().Unix()),
		}
		err := msg.Save()
		if err != nil {
			log.Printf("消息%+v保存失败：%s\n", msg, err.Error())
			ret[i] = false
		} else {
			ret[i] = true
		}
	}
	//返回每个是成功还是失败
	return ret
}

// 对参数做各种验证
func checkParams(singleList batchRequest, fromIp string) error {
	for _, single := range singleList {
		//验证参数
		if err := validateSingleRequest(single); err != nil {
			return err
		}
		//验证来源
		if err := checkProduct(single, fromIp); err != nil {
			return err
		}
		//验证cmd
		if err := checkCommand(single); err != nil {
			return err
		}
	}
	return nil
}

//检查cmd和project是否匹配并且已经提前定义过
func checkCommand(request singleRequest) error {
	for _, itemCmd := range common.Config.CommandList {
		if itemCmd.Project == request.Project && itemCmd.Command == request.Cmd {
			//找到规定的命令了
			return nil
		}
	}
	return common.ThrowNotice(common.ErrorCodeUnknowCommand, errors.New("存在未知的cmd"))
}

// 验证来源
func checkProduct(request singleRequest, ip string) error {
	for _, product := range common.Config.ProductList {
		if product.Project == request.Project {
			// 找到了来源 判断ip
			for _, ipa := range product.AllowIp {
				if ipa == ip {
					// ip也合法
					return nil
				}
			}
		}
	}
	return common.ThrowNotice(common.ErrorCodeUnknowProduct, errors.New("不合法的消息来源"))
}

//输出错误信息
func responseWithError(writer http.ResponseWriter, err error) {
	resp := responseBody{
		Code:    0,
		Message: err.Error(),
		Data:    nil,
	}

	switch e := err.(type) {
	case common.Notice:
		//自定义的Error类型
		log.Printf("notice error :[%d] %s\n", e.Code(), e.Error())
		resp.Code = uint(e.Code())
		responseWithJson(writer, resp)
	default:
		log.Printf("error : %s\n", e.Error())
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
		log.Println("response error: " + err.Error())
	}
}
