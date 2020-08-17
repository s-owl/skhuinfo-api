package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var MealHttpMock HttpClient = &HttpMock{
	map[string]string{
		MEAL_LIST: "test/meal_list.html",
	},
}

func Test_getMealId(t *testing.T) {
	assert := assert.New(t)
	if list, err := getMealID(MealHttpMock); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(
			MealID{
				ID:    389,
				Title: "학생식당 주간메뉴입니다(12/2-12/6)",
				Date:  "2019-11-29",
			},
			list[0],
			"getMealID 반환값 테스트",
		)
	}
}
