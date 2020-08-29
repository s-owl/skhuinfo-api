package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var SchedulesHTTPMock HttpClient = &HttpMock{
	map[string]string{
		SchedulesURL: "test/schedules_current.html",
		SchedulesURL + "?strYear=2020&strMonth=9": "test/schedules_202009.html",
	},
}

func Test_getScheduleData(t *testing.T) {
	assert := assert.New(t)
	if list, err := getScheduleData(SchedulesHTTPMock, SchedulesURL); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(
			ScheduleItem{
				Period:  "07-30 ~ 08-02",
				Content: "수강바구니신청기간",
			},
			list[0],
			"getScheduleData 반환값 테스트(인자 없이 실행)",
		)
	}
	if list, err := getScheduleData(SchedulesHTTPMock, SchedulesURL+"?strYear=2020&strMonth=9"); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(
			ScheduleItem{
				Period:  "08-31 ~ 09-04",
				Content: "2차 수강신청 및 변경기간",
			},
			list[0],
			"getScheduleData 반환값 테스트(년도, 월 인자 있음)",
		)
	}
}
