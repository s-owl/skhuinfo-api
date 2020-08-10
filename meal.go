package main

import (
	"errors"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

// 학식 URL을 저장하기 위한 구조체
type MealURL struct {
	Title string `json:"title"`
	Url   string `json:"url"`
	Date  string `json:"date"`
}

/*
1주간의 학식이 담겨있는 링크들을 게시판에서 추출한다.
MealURL 구조체 배열로 반환한다.
*/
func getMealURL() (list []MealURL, err error) {
	// url 목록을 저장할 변수
	list = []MealURL{}
	defer WhereInError(&err, "학식 목록")

	// 학식이 있는 게시판 목록을 가져온다
	client := &http.Client{Timeout: TIMEOUT}
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
		list = append(list, MealURL{
			Title: item.Children().Eq(1).Find("a").Text(),
			Url:   MEAL_BOARD + item.Children().Eq(1).Find("a").AttrOr("href", ""),
			Date:  item.Children().Eq(3).Text(),
		})
	})

	return list, nil
}

/*
1주간의 학식이 담겨있는 링크들을 게시판에서 추출하고 바로 JSON 형태로 전송한다.
이전에 SKHUS 프로젝트 사용했던 방식으로 여기에서 얻은 링크를
다른 API에 인수로 제공해서 학식 정보를 얻어냈다.

{
	"data": [
		{
			"url": "http://skhu.ac.kr/uni_zelkova/...",
			"title": "학생식당 주간메뉴...",
			"date": "2019-XX-XX",
		}, ...
	],
	"error": "error message when occured, else empty"
}
*/
func GetMealIds(c *gin.Context) {
	// 게시판에서 학식 목록을 가져오려 시도한다.
	list, err := getMealURL()
	if err != nil {
		// 에러 종류를 파악할 수 있는 구조체를 선언
		var wrap *APIError
		// 반드시 있는 에러 타입으로 if로 굳이 확인하지 않아도 된다.
		errors.As(err, &wrap)
		// 발생된 에러로 http 상태 번호를 활당한다.
		statusCode := wrap.GetHttpStatusCode()

		c.JSON(statusCode, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": list,
	})
}
