package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

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
// @Failure 404 {object} ErrorMessage
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
