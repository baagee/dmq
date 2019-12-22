package handler

import (
	"encoding/json"
	"errors"
	"github.com/baagee/dmq/common"
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
func save(singleList batchRequest) []interface{} {
	length := len(singleList)
	// 切片需要make
	ret := make([]interface{}, length)
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
		// 验证消息是否重复
		// 获取消息hash 查询redis判断是否存在
		if mId := msg.CheckExists(); mId == 0 {
			// 不存在 就保存
			err := msg.Save()
			if err != nil {
				common.RecordError(err)
				ret[i] = false //失败返回false
				continue
			} /*else{
				// nothing
			}*/
		} else {
			//msg.Id = mId //返回已存在的消息ID
		}
		ret[i] = msg.Id //返回消息ID
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
	cmd, exists := common.Config.CommandMap[common.GetConfigCmdKey(request.Cmd)]
	if exists == false {
		return common.ThrowNotice(common.ErrorCodeUnknowCommand, errors.New("存在未知的cmd"))
	}
	if cmd.Project != request.Project {
		return common.ThrowNotice(common.ErrorCodeUnknowCommand, errors.New("不匹配的cmd和project"))
	}
	return nil
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
