package main

import (
	"io"
	"io/ioutil"
	"strings"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

func EucKrReaderToUtf8Reader(body io.Reader) io.Reader {
	rInUTF8 := transform.NewReader(body, korean.EUCKR.NewDecoder())
	decBytes, _ := ioutil.ReadAll(rInUTF8)
	decrypted := string(decBytes)
	return strings.NewReader(decrypted)
}
