package main

import (
	"fmt"
	"net/http"
)

// 에러 분류를 위한 코드
type ErrorCode int

// 에러 분류
const (
	NetworkError ErrorCode = iota + 1
	EncodingError
)

// 에러 코드와 실제 에러를 포장하는 구조체
type APIError struct {
	Code  ErrorCode
	inner error
}

// 에러코드에서 API 에러를 생성한다.
func (code ErrorCode) CreateError(inner error) error {
	return &APIError{
		code,
		inner,
	}
}

// 에러의 문자열 표현
func (e *APIError) Error() string {
	message := "알수없는 오류"
	switch e.Code {
	case NetworkError:
		message = "네트워크 오류"
	case EncodingError:
		message = "인코딩 오류"
	}

	if e.inner != nil {
		message = message + ": " + e.inner.Error()
	}

	return message
}

// 에러에서 http 상태 코드를 알아낸다.
func (e *APIError) GetHttpStatusCode() int {
	statusCode := http.StatusInternalServerError
	switch e.Code {
	case NetworkError:
		statusCode = http.StatusNotFound
	case EncodingError:
		statusCode = http.StatusBadGateway
	}

	return statusCode
}

/*
해당 에러가 어디에서 일어났는지 표시해주는 함수
명명된 반환 변수(named return value)로 error 타입을 만든 후 defer로 사용해주세요.
ex)
func blahblahfunc() (err error) {
	defer WhereInError(err, "아무튼")

	if err = somethingError() {
		return
	}

	return
}
*/
func WhereInError(err *error, where string) {
	if *err != nil {
		*err = fmt.Errorf("%s 오류: %w", where, *err)
	}
}
