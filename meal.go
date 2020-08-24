package main

import (
	"net/http"
	"regexp"
	"strconv"

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
