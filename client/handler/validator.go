package handler

import (
	"errors"
	"github.com/baagee/dmq/common"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	validator9 "gopkg.in/go-playground/validator.v9"
	zhTranslations "gopkg.in/go-playground/validator.v9/translations/zh"
	"log"
)

// 验证参数
func validateSingleRequest(single singleRequest) error {
	//中文翻译器
	zhCh := zh.New()
	uni := ut.New(zhCh)
	trans, _ := uni.GetTranslator("zh")
	validator := validator9.New()
	//注册中文翻译
	errt := zhTranslations.RegisterDefaultTranslations(validator, trans)
	if err := validator.Struct(single); err != nil {
		for _, itemErr := range err.(validator9.ValidationErrors) {
			if errt != nil {
				log.Println("RegisterDefaultTranslations " + errt.Error())
				return common.Notice{
					CodeInt: common.ErrorCodeValidateFailed,
					Err:     err,
				}
			} else {
				return common.Notice{
					CodeInt: common.ErrorCodeValidateFailed,
					Err:     errors.New(itemErr.Translate(trans)),
				}
			}
		}
	}
	return nil
}