package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func somethingError() error {
	return errors.New("테스트")
}

func Test_APIErrorFromErrorCode(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(
		UnknownError.CreateError(nil).Error(),
		"알수없는 오류",
		"UnknownError's Message")
	assert.Equal(
		UnknownError.CreateError(somethingError()).Error(),
		"알수없는 오류: 테스트",
		"inner Error print test")
}

func Test_MakeErrorMessage(t *testing.T) {
	assert := assert.New(t)

	err := UnknownError.CreateError(somethingError())
	message := MakeErrorMessage(err)

	assert.Equal(message.Message, "알수없는 오류: 테스트", "ErrorMessage message test")

	err = NetworkError.CreateError(somethingError())
	message = MakeErrorMessage(err)

	assert.Equal(message.StatusCode, 404, "Check StatusCode")
}

func Test_WhereIsError(t *testing.T) {
	assert := assert.New(t)

	err := somethingError()
	WhereInError(&err, "아무튼")
	assert.Equal(err.Error(), "아무튼 오류: 테스트")
}
