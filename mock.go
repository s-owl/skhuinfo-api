package main

import (
	"errors"
	"net/http"
	"os"
)

// 실제 http.Client 구조체와 Mocking용 구조체와 같은 타입으로 활용하기 위한 인터페이스
type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// 실제 요청을 위한 http.Client를 생성하는 함수
func HttpReal() HttpClient {
	return &http.Client{Timeout: TIMEOUT}
}

// http 요청을 해석하여 맵에 해당 http 주소와 그에 해당하는 파일 주소가 존재하면 그 파일을 응답의 Body로 반환한다.
type HttpMock struct {
	UrlToFile map[string]string
}

// http 요청을 해석하고 그 요청에 해당하는 파일을 응답의 Body로 반환한다.
func (mock *HttpMock) Do(req *http.Request) (res *http.Response, err error) {
	defer WhereInError(&err, "테스트 요청")

	// 요청한 주소가 맵에 없으면 오류 발생
	filepath, exist := mock.UrlToFile[req.URL.String()]
	if !exist {
		err = errors.New("없는 테스트 주소: " + req.URL.String())
		return
	}

	// 해당 요청으로 보내야 할 파일을 불러온다.
	file, err := os.Open(filepath)
	if err != nil {
		return
	}

	// 불러온 파일을 응답의 Body로 넣어서 반환한다.
	res = &http.Response{
		Body: file,
	}
	return
}
