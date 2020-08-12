package main

import (
	"time"
)

const (
	// 시간 제한 초
	TIMEOUT = 5 * time.Second

	// 학교 홈페이지 주소
	SKHU_URL = "http://skhu.ac.kr/"
	// 학식 목록 주소
	MEAL_LIST = SKHU_URL + "uni_zelkova/uni_zelkova_4_3_list.aspx"
	// 학식 게시판 주소
	MEAL_BOARD = SKHU_URL + "uni_zelkova/uni_zelkova_4_3_view.aspx?idx="
)
