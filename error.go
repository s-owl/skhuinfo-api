package main

import (
	"errors"
	"fmt"
	"net/http"
)

// 에러 분류를 위한 코드
type ErrorCode int

// 에러 분류
const (
	UnknownError ErrorCode = iota
	NetworkError
	EncodingError
	NotFoundError
	InvalidError
)

// 에러 코드와 실제 에러를 포장하는 구조체
type APIError struct {
	code  ErrorCode
	inner error
}

// 에러코드에서 API 에러를 생성한다.
func (code ErrorCode) CreateError(inner error) *APIError {
	return &APIError{
		code,
		inner,
	}
}

// APIError의 Unwrap 기능 구현
func (e *APIError) Unwrap() error {
	return e.inner
}

// 에러의 문자열 표현
func (e *APIError) Error() string {
	message := "알수없는 오류"
	switch e.code {
	case NetworkError:
		message = "네트워크 오류"
	case EncodingError:
		message = "인코딩 오류"
	case NotFoundError:
		message = "없는 자료"
	case InvalidError:
		message = "잘못된 인수 오류"
	}

	if e.inner != nil {
		message = message + ": " + e.inner.Error()
	}

	return message
}

// 에러에서 http 상태 코드를 알아낸다.
func (e *APIError) getHttpStatusCode() int {
	// 해당 목록에 없으면 기본적으로 500 서버 내부 오류
	statusCode := http.StatusInternalServerError

	switch e.code {
	case NetworkError:
		fallthrough
	case EncodingError:
		statusCode = http.StatusBadGateway
	case NotFoundError:
		statusCode = http.StatusNotFound
	case InvalidError:
		statusCode = http.StatusBadRequest
	}

	return statusCode
}

// 에러를 JSON으로 전송하기 위한 타입
type ErrorMessage struct {
	StatusCode int    `json:"-"`
	Message    string `json:"error"`
}

/*
에러를 받아 JSON형태 바꾸는 함수
*/
func MakeErrorMessage(err error) ErrorMessage {
	// 에러를 분석하고 전송하기 위한 변수
	var wrap *APIError
	var statusCode int
	var message string

	// APIError 형식인지 확인하고 아니면 APIError 형식으로 변경한다.
	if !errors.As(err, &wrap) {
		wrap = UnknownError.CreateError(err)
		err = wrap
	}

	// 필요한 자료를 받아온다.
	statusCode = wrap.getHttpStatusCode()
	message = err.Error()

	// JSON 형식으로 변환한다.
	return ErrorMessage{
		statusCode,
		message,
	}
}

/*
해당 에러가 어디에서 일어났는지 표시해주는 함수
명명된 반환 변수(named return value)로 error 타입을 만든 후 defer로 사용해주세요.
ex)
func blahblahfunc() (err error) {
	defer WhereInError(&err, "아무튼")

	if err = somethingError(); err != nil {
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
