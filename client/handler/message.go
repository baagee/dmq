package handler

import (
	"encoding/json"
	"errors"
	"github.com/baagee/dmq/common"
	"log"
	"net/http"
	"strconv"
	"time"
)

//当个请求的结构体
type singleRequest struct {
	Cmd       string `json:"cmd" validate:"required"`     // 命令点
	Timestamp uint64 `json:"timestamp"`                   // 执行时间
	Params    string `json:"params"`                      // 命令参数
	RequestId string `json:"request_id"`                  // 请求ID
	Project   string `json:"project" validate:"required"` // 项目
	Bucket    string `json:"bucket" validate:"required"`  // 消息桶
}

//批量请求结构体
type batchRequest []singleRequest

//获取消息状态
func MessageStatus(writer http.ResponseWriter, request *http.Request) {
	msgId := request.URL.Query().Get("msg_id")
	msgIdInt, err := strconv.Atoi(msgId)
	if err != nil {
		responseWithError(writer, common.ThrowNotice(common.ErrorCodeParseParamsFailed, errors.New("msgId为空")))
		return
	}
	if msgIdInt == 0 {
		responseWithError(writer, common.ThrowNotice(common.ErrorCodeParseParamsFailed, errors.New("不是合法的msgId")))
		return
	}
	msg := common.Message{
		Id: uint64(msgIdInt),
	}

	ret, err1 := msg.Status()
	if err1 != nil {
		responseWithError(writer, common.ThrowNotice(common.ErrorCodeGetStatusFailed, err))
		return
	}
	var respBody responseBody
	respBody.Data = ret
	responseWithJson(writer, respBody)
}

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

	//保存
	var respBody responseBody
	respBody.Data = save(singleList, common.GetClientIP(request))[0]
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

	//保存
	var respBody responseBody
	respBody.Data = save(singleList, common.GetClientIP(request))
	responseWithJson(writer, respBody)
}

//保存命令
func save(singleList batchRequest, fromIp string) []interface{} {
	length := len(singleList)
	// 切片需要make
	ret := make([]interface{}, length)
	idBaseNumber := common.GetIdBaseNumber(int64(length))
	for i, single := range singleList {
		// 	验证参数
		err := checkParams(single, fromIp)
		if err != nil {
			//参数验证失败 返回 错误信息
			ret[i] = err.Error()
			continue
		}
		if single.Timestamp == 0 {
			// 时间为0 表示当前时间 立即执行
			single.Timestamp = uint64(time.Now().Unix())
		}
		msg := common.Message{
			Cmd:        single.Cmd,
			Timestamp:  single.Timestamp,
			Params:     single.Params,
			Project:    single.Project,
			Bucket:     single.Bucket,
			RequestId:  single.RequestId,
			CreateTime: uint64(time.Now().Unix()),
		}
		// 验证消息是否重复
		// 获取消息hash 查询redis判断是否存在
		if mId := msg.CheckExists(); mId == 0 {
			// 不存在 就保存 获取msgId
			msg.Id = common.GenerateId(int64(i), idBaseNumber, int64(length))
			err := msg.Save()
			if err != nil {
				common.RecordError(err)
				ret[i] = false //消息保存失败 返回false
				continue
			}
		}
		ret[i] = msg.Id //消息保存成功 返回消息ID
		log.Printf("message: %+v\n", msg)
	}
	//返回每个是成功还是失败
	return ret
}

// 对参数做各种验证
func checkParams(single singleRequest, fromIp string) error {
	//验证参数
	if err := validateSingleMessageRequest(single); err != nil {
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
	return nil
}

//检查cmd和project是否匹配并且已经提前定义过
func checkCommand(request singleRequest) error {
	cmd, exists := common.Config.CommandMap[common.GetConfigCmdKey(request.Cmd)]
	if exists == false {
		return common.ThrowNotice(common.ErrorCodeUnknownCommand, errors.New("存在未知的cmd"))
	}
	if cmd.Project != request.Project {
		return common.ThrowNotice(common.ErrorCodeUnknownCommand, errors.New("不匹配的cmd和project"))
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
	return common.ThrowNotice(common.ErrorCodeUnknownProduct, errors.New("不合法的消息来源"))
}
