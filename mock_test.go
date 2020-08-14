package main

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TEST_PATH = SKHU_URL + "test"
)

// HttpMock을 테스트한다.
func Test_HttpMock(t *testing.T) {
	assert := assert.New(t)

	mocking := &HttpMock{
		map[string]string{
			TEST_PATH: "test/mock.html",
		},
	}

	var req *http.Request
	var buf *strings.Builder
	req, _ = http.NewRequest("GET", TEST_PATH, nil)
	test_html, err := mocking.Do(req)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer test_html.Body.Close()
	buf = new(strings.Builder)
	io.Copy(buf, test_html.Body)
	assert.Equal("Hello, Mock!\n", buf.String(), "Read File Test")
}
