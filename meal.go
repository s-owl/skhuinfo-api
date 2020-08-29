package main

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

// 학식 URL을 저장하기 위한 구조체
type MealID struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Date  string `json:"date"`
}

// URL에서 ID를 추출하기 위한 정규표현식
var regexIDFromURL *regexp.Regexp = regexp.MustCompile("idx=(\\d+)&")

/*
1주간의 학식이 담겨있는 링크들을 게시판에서 추출한다.
MealID 구조체 배열로 반환한다.
*/
func getMealID(client HttpClient) (list []MealID, err error) {
	// url 목록을 저장할 변수
	list = []MealID{}
	defer WhereInError(&err, "학식 목록")

	// 학식이 있는 게시판 목록을 가져온다
	req, err := http.NewRequest("GET", MEAL_LIST, nil)
	res, err := client.Do(req)
	if err != nil {
		err = NetworkError.CreateError(err)
		return
	}
	defer res.Body.Close()

	// euckr에서 utf8로 인코딩 변환을 시도한다.
	doc, err := goquery.NewDocumentFromReader(EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		return nil, EncodingError.CreateError(err)
	}

	// css를 기준으로 글 목록을 분류
	doc.Find("table.board_list > tbody > tr").Each(func(i int, item *goquery.Selection) {
		// url에서 id를 파싱한다.
		// url 파싱에 실패하면 무시한다.
		url := item.Children().Eq(1).Find("a").AttrOr("href", "")
		plainID := regexIDFromURL.FindStringSubmatch(url)
		if len(plainID) != 2 {
			return
		}
		id, _ := strconv.Atoi(plainID[1])

		list = append(list, MealID{
			Title: item.Children().Eq(1).Find("a").Text(),
			ID:    id,
			Date:  item.Children().Eq(3).Text(),
		})
	})

	return
}

type GetMealIdsResult struct {
	Data []MealID `json:"data"`
}

// GetMealIds godoc
// @Summary 학식 게시판에서 학식 목록을 가져온다.
// @Description MealID 배열인 data를 가진 구조체를 리턴받는다.
// @Produce json
// @Success 200 {object} GetMealIdsResult
// @Failure 502 {object} ErrorMessage
// @Router /meal/ids [get]
func GetMealIds(c *gin.Context) {
	// 게시판에서 학식 목록을 가져오려 시도한다.
	list, err := getMealID(HttpReal())
	// 에러가 있으면 에러를 전송하고 끝내고 아니면 데이터를 전송
	if err != nil {
		message := MakeErrorMessage(err)
		c.JSON(message.StatusCode, message)
		return
	}

	c.JSON(http.StatusOK, GetMealIdsResult{
		list,
	})
}

// 실제 식단 데이터
type Diet struct {
	Diet    string `json:"diet"`
	Calorie string `json:"calorie"`
}

// 점심 학식
type Lunch struct {
	A Diet `json:"a"`
	B Diet `json:"b"`
	C Diet `json:"c"`
}

// 저녁 학식
type Dinner struct {
	A Diet `json:"a"`
}

// 학식 데이터
type MealData struct {
	Day    string `json:"day"`
	Date   string `json:"date"`
	Lunch  Lunch  `json:"lunch"`
	Dinner Dinner `json:"dinner"`
}

// css 선택자 상수
const theadSelector string = `thead > tr:nth-child(%d) > th:nth-child(%d)`
const tbodySelector string = `tbody > tr:nth-child(%d) > td:nth-child(%d)`

// 학식 품목을 표에서 추출하는 함수
func processDietData(sel *goquery.Selection, trIndex int, tdIndex int) Diet {
	item := sel.Find(fmt.Sprintf(tbodySelector, trIndex, tdIndex))
	htmlContent, _ := item.Html()
	content := strings.ReplaceAll(htmlContent, "<br/>", "\n")
	calorie := sel.Find(fmt.Sprintf(tbodySelector, trIndex+1, tdIndex)).Text()
	return Diet{
		content,
		calorie,
	}
}

// 게시판 ID를 통해 한 주의 학식을 가져온다.
func getMealDataWithID(client HttpClient, id int) (week []MealData, err error) {
	defer WhereInError(&err, "식단 데이터 처리")

	// 요청 생성
	req, _ := http.NewRequest("GET", MEAL_BOARD+strconv.Itoa(id), nil)
	res, err := client.Do(req)
	if err != nil {
		err = NetworkError.CreateError(err)
		return
	}
	defer res.Body.Close()

	// EucKr to UTF-8
	doc, err := goquery.NewDocumentFromReader(EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		err = EncodingError.CreateError(err)
		return
	}

	// 학식 표 찾기
	mealTable := doc.Find("table.cont_c")

	// 학식 표 순회
	for i := 0; i < 5; i++ {
		// 학식을 추출 후 배열에 추가
		week = append(week, MealData{
			Day:  mealTable.Find(fmt.Sprintf(theadSelector, 1, i+2)).Text(),
			Date: mealTable.Find(fmt.Sprintf(theadSelector, 2, i+3)).Text(),
			Lunch: Lunch{
				A: processDietData(mealTable, 1, i+3),
				B: processDietData(mealTable, 3, i+2),
				C: processDietData(mealTable, 5, i+2),
			},
			Dinner: Dinner{
				A: processDietData(mealTable, 7, i+3),
			},
		})
	}

	return
}

var regexDateFromTitle *regexp.Regexp = regexp.MustCompile("(\\d{1,2})\\D{1,2}(\\d{1,2})\\D{1,2}(\\d{1,2})\\D{1,2}(\\d{1,2})\\D?")

// 현재 시간을 기준으로 해당 주의 시간표를 추출한다.
func getMealDataWithTime(client HttpClient, now time.Time) (week []MealData, err error) {
	defer WhereInError(&err, "현재 주간 학식")

	weekday := now.Weekday()
	if weekday < time.Monday || weekday > time.Friday {
		err = NotFoundError.CreateError(errors.New("주말에는 학식이 제공하지 않습니다."))
		return
	}

	// 학식 게시판 목록을 가져온다.
	ids, err := getMealID(client)
	if err != nil {
		return
	}

	// 시간대 설정
	loc, _ := time.LoadLocation("Asia/Seoul")

	for _, id := range ids {
		// 올린 날짜에서 년도를 가져오고
		year, _ := strconv.Atoi(strings.Split(id.Date, "-")[0])

		// 제목에서 기간을 가져온다.
		parsedDuration := regexDateFromTitle.FindStringSubmatch(id.Title)
		if len(parsedDuration) != 5 {
			continue
		}

		// 가져온 기간을 숫자 배열로 다시 변환한다.
		numberDuration := []int{}
		for _, plain := range parsedDuration[1:] {
			num, _ := strconv.Atoi(plain)
			numberDuration = append(numberDuration, num)
		}

		// 년도와 기간을 통해 시작일과 종료일을 Time 자료형으로 생성한다.
		startDay := time.Date(year, time.Month(numberDuration[0]), numberDuration[1], 0, 0, 0, 0, loc)
		endDay := time.Date(year, time.Month(numberDuration[2]), numberDuration[3], 23, 59, 59, 99, loc)

		// 현재 시간이 시작일과 종료일 사이에 있는지 확인한다.
		if endDay.After(now) && startDay.Before(now) {
			// 그 주의 학식을 가져온다.
			week, err = getMealDataWithID(client, id.ID)
			if err != nil {
				return
			}

			return
		}
	}

	err = NotFoundError.CreateError(errors.New("해당 기간에 학식이 존재하지 않습니다."))
	return
}

// 현재 시간을 기준으로 해당 요일의 시간표를 추출한다.
func getMealDataWithWeekday(client HttpClient, now time.Time, weekday int) (day []MealData, err error) {
	defer WhereInError(&err, "요일별 학식")

	if weekday < 1 || weekday > 5 {
		err = NotFoundError.CreateError(errors.New("주말에는 학식이 제공하지 않습니다."))
		return
	}

	data, err := getMealDataWithTime(client, now)
	if err != nil {
		return
	}

	day = []MealData{
		data[weekday-1],
	}
	return
}

type GetMealDataResult struct {
	Data []MealData `json:"data"`
}

// godoc GetMealData
// @Summary 학식 게시판에서 학식을 가져온다.
// @Description MealData 배열인 data를 가진 구조체를 리턴받는다.
// @Produce json
// @Param id query int false "게시물 ID" 382
// @Param day query int false "날짜 요일 (0~5)" 5
// @Success 200 {object} GetMealDataResult
// @Failure 400 {object} ErrorMessage
// @Failure 404 {object} ErrorMessage
// @Failure 502 {object} ErrorMessage
// @Router /meal/get [get]
func GetMealData(c *gin.Context) {
	// GET 인수를 받는다. 기본값은 문자열 "0"
	plainId := c.DefaultQuery("id", "0")
	plainDay := c.DefaultQuery("day", "0")

	// GET 인수로 받은 변수들을 전부 숫자로 변경한다.
	id, err1 := strconv.Atoi(plainId)
	day, err2 := strconv.Atoi(plainDay)
	if err1 != nil || err2 != nil {
		err := InvalidError.CreateError(errors.New("정수로만 사용해주세요"))
		message := MakeErrorMessage(err)
		c.JSON(message.StatusCode, message)
		return
	}

	// 함수의 실행 결과를 확인하고 데이터를 전송하는 함수
	sendMealData := func(data []MealData, err error) {
		if err != nil {
			message := MakeErrorMessage(err)
			c.JSON(message.StatusCode, message)
			return
		}

		c.JSON(http.StatusOK, GetMealDataResult{data})
	}

	if id != 0 && day != 0 {
		data, err := getMealDataWithID(HttpReal(), id)
		if len(data) == 5 {
			data = []MealData{
				data[day-1],
			}
		}
		sendMealData(data, err)
	} else if day != 0 {
		now := time.Now()
		data, err := getMealDataWithWeekday(HttpReal(), now, day)
		sendMealData(data, err)
	} else if id != 0 {
		data, err := getMealDataWithID(HttpReal(), id)
		sendMealData(data, err)
	} else {
		now := time.Now()
		data, err := getMealDataWithTime(HttpReal(), now)
		sendMealData(data, err)
	}
}
