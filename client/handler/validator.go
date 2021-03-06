package handler

import (
	"errors"
	"github.com/baagee/dmq/common"
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	validator9 "gopkg.in/go-playground/validator.v9"
	zhTranslations "gopkg.in/go-playground/validator.v9/translations/zh"
)

var (
	zhCh      locales.Translator
	uni       *ut.UniversalTranslator
	trans     ut.Translator
	validator *validator9.Validate
	errt      error
)

func init() {
	zhCh = zh.New()
	uni = ut.New(zhCh)
	trans, _ = uni.GetTranslator("zh")
	validator = validator9.New()
	//注册中文翻译
	errt = zhTranslations.RegisterDefaultTranslations(validator, trans)
	if errt != nil {
		common.RecordError(errt)
	}
}

// 验证提交消息的参数
func validateSingleMessageRequest(singleMessage singleRequest) error {
	if err := validator.Struct(singleMessage); err != nil {
		for _, itemErr := range err.(validator9.ValidationErrors) {
			if errt != nil {
				common.RecordError(err)
				return common.ThrowNotice(common.ErrorCodeValidateFailed, err)
			} else {
				return common.ThrowNotice(common.ErrorCodeValidateFailed, errors.New(itemErr.Translate(trans)))
			}
		}
	}
	return nil
}
