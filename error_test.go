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
		"알수없는 오류",
		UnknownError.CreateError(nil).Error(),
		"UnknownError's Message")
	assert.Equal(
		"알수없는 오류: 테스트",
		UnknownError.CreateError(somethingError()).Error(),
		"inner Error print test")
}

func Test_MakeErrorMessage(t *testing.T) {
	assert := assert.New(t)

	err := UnknownError.CreateError(somethingError())
	message := MakeErrorMessage(err)

	assert.Equal(
		"알수없는 오류: 테스트",
		message.Message,
		"ErrorMessage message test",
	)

	err = NetworkError.CreateError(somethingError())
	message = MakeErrorMessage(err)

	assert.Equal(message.StatusCode, 404, "Check StatusCode")
}

func Test_WhereIsError(t *testing.T) {
	assert := assert.New(t)

	err := somethingError()
	WhereInError(&err, "아무튼")
	assert.Equal(
		"아무튼 오류: 테스트",
		err.Error(),
		"WhereInError message append test",
	)
}
