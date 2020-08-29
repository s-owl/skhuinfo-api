package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var MealHttpMock HttpClient = &HttpMock{
	map[string]string{
		MEAL_LIST:          "test/meal_list.html",
		MEAL_BOARD + "389": "test/meal_board_389.html",
		MEAL_BOARD + "380": "test/meal_board_380.html",
	},
}

func Test_getMealId(t *testing.T) {
	assert := assert.New(t)
	if list, err := getMealID(MealHttpMock); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(
			mealID{
				ID:    389,
				Title: "학생식당 주간메뉴입니다(12/2-12/6)",
				Date:  "2019-11-29",
			},
			list[0],
			"getMealID 반환값 테스트",
		)
	}
}

func Test_getMealDataWithID(t *testing.T) {
	assert := assert.New(t)
	if week, err := getMealDataWithID(MealHttpMock, 389); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(
			diet{
				`사골순대국

쌀밥
미트볼조림
감자매콤조림
두부구이
김치`,
				"",
			},
			week[0].Lunch.A,
			"getMealDataWithID 반환값 테스트",
		)
	}
}

func Test_getMealDataWithWeekday(t *testing.T) {
	assert := assert.New(t)
	loc, _ := time.LoadLocation("Asia/Seoul")
	now := time.Date(2019, 10, 8, 0, 0, 0, 0, loc)
	if week, err := getMealDataWithWeekday(MealHttpMock, now, 3); err != nil {
		t.Fatal(err)
	} else {
		assert.Equal(
			diet{
				"한글날",
				"",
			},
			week[0].Lunch.A,
			"getMealDataWithWeekDay 반환값 테스트",
		)
	}
}
