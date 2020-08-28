package main

import (
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

// 현재 시간을 기준으로 해당 요일의 시간표를 추출한다.
func getMealDataWithWeekNum(client HttpClient, now *time.Time, weeknum int) (day []MealData, err error) {
	defer WhereInError(&err, "요일별 학식")

	// 현재 주말인 경우를 제외
	if now.Weekday() < 1 || now.Weekday() > 5 {
		err = NotFoundError.CreateError(nil)
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
		if startDay.After(*now) && endDay.Before(*now) {
			// 현재 주간 식단표를 가져와서 해당 요일의 식단만 분리한다.
			var data []MealData
			data, err = getMealDataWithID(client, id.ID)
			if err != nil {
				return
			}

			day = []MealData{
				data[weeknum-1],
			}
			return
		}
	}

	err = NotFoundError.CreateError(nil)
	return
}
