package main

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

// GetSchedulesResult - 월간 학사 일정 결과 데이터 구조체
type GetSchedulesResult struct {
	Schedules []ScheduleItem `json:"schedules"`
}

// GetSchedules godoc
// @Summary 월간 학사 일정 조회
// @Description ScheduleItem 배열인 schedules를 가진 구조체를 리턴받는다.
// @Param year path int false "년도"
// @Param month path int false "월"
// @Produce json
// @Success 200 {object} GetSchedulesResult
// @Failure 404 {object} ErrorMessage
// @Failure 502 {object} ErrorMessage
// @Router /schedules/{year}/{month} [get]
func GetSchedules(c *gin.Context) {
	year := c.Param("year")
	month := c.Param("month")
	targetURL := SCHEDULES_URL
	if year != "" || month != "" {
		targetURL = SCHEDULES_URL + fmt.Sprintf("?strYear=%s&strMonth=%s", year, month)
	}
	list, err := getScheduleData(HttpReal(), targetURL)
	// 에러가 있으면 에러를 전송하고 끝내고 아니면 데이터를 전송
	if err != nil {
		message := MakeErrorMessage(err)
		c.JSON(message.StatusCode, message)
		return
	}
	c.JSON(http.StatusOK, GetSchedulesResult{
		Schedules: list,
	})
}

// ScheduleItem 은 학사 일정 데이터 항목 구조체 입니다.
type ScheduleItem struct {
	Period  string `json:"period"`
	Content string `json:"content"`
}

func getScheduleData(client HttpClient, targetURL string) ([]ScheduleItem, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	res, err := client.Do(req)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		return nil, EncodingError.CreateError(err)
	}

	schedules := []ScheduleItem{}
	doc.Find("div.info > table > tbody > tr").Each(func(i int, item *goquery.Selection) {
		if i > 0 {
			schedules = append(schedules, ScheduleItem{
				Period:  item.Children().Eq(0).Text(),
				Content: item.Children().Eq(1).Text(),
			})
		}
	})
	return schedules, nil
}
