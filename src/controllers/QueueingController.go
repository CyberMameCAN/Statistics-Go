package controllers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

type QueueingTheory struct {
	Window   string `form:"window" binding:"required,min=1,max=1"`
	Arrived  string `form:"arrived" binding:"required,min=1,max=3"`
	Servises string `form:"servises" binding:"required,min=1,max=3"`
}

type QueueingResult struct {
	WaitMan        string
	WaitTime       string
	TurnAroundTime string
}

var queueing_result QueueingResult // 計算結果を格納

func init() {
	initOutputData()
}

func initOutputData() {
	queueing_result.WaitMan = ""
	queueing_result.WaitTime = ""
	queueing_result.TurnAroundTime = ""
}

func (form *QueueingTheory) message() map[string]map[string]string {
	return map[string]map[string]string{
		"Window": {
			"required": "窓口数は必須です",
			"max":      "窓口数は1桁以下です",
		},
		"Arrived": {
			"required": "平均到着時間は必須です",
			"max":      "平均到着時間は1桁以下です",
		},
		"Servises": {
			"required": "平均サービス時間は必須です",
			"max":      "平均サービス時間は1桁以下です",
		},
	}
}

func Queueing(c *gin.Context) {
	var kekka string

	// 初期化
	initOutputData()
	onDisp := false // true 表示する, false しない

	// フォームよりデータを取得
	window := c.PostForm("window")
	arrived := c.PostForm("arrived")
	servises := c.PostForm("servises")

	var form QueueingTheory
	// form := forms.User{
	// 	Name:      r.FormValue("name"),
	// 	FirstName: r.FormValue("firstName"),
	// 	LastName:  r.FormValue("lastName"),
	// 	Email:     r.FormValue("email"),
	// }
	// Bindの代表的２つ　MustBindWith, ShouldBind
	if err := c.ShouldBind(&form); err != nil {
		log.Println("formの取得結果：", form)

		if ve, ok := err.(validator.ValidationErrors); ok {
			// リクエストが間違っている時の処理
			fmt.Println("バリデーションエラー内容：", ve)
			c.HTML(http.StatusOK, "index.html", gin.H{
				"form":  form,
				"kekka": queueing_result,
				"logs":  ve,
			})
			return
		} else {
			// pass
		}
		log.Println(err)
	}

	window_f, _ := strconv.ParseFloat(window, 64)
	arrived_f, _ := strconv.ParseFloat(arrived, 64)
	servises_f, _ := strconv.ParseFloat(servises, 64)
	log.Println("計算準備データ：", window_f, arrived_f, servises_f)
	// 取得データのチェック
	if window_f == 0 || arrived_f == 0 || servises_f == 0 {
		param_map := map[string]float64{"窓口数": window_f, "平均待ち時間": arrived_f, "平均サービス時間": servises_f}
		kekka = "パラメータエラー： "
		for key, val := range param_map {
			if val == 0 {
				kekka = kekka + fmt.Sprintf("%s: %.0f", key, val)
			}
		}
		c.HTML(http.StatusOK, "index.html", gin.H{
			"form":  form,
			"kekka": queueing_result,
			"logs":  kekka,
		})
		return
	}
	//

	// Data is OK.

	err := exec_queueing(arrived_f, servises_f, window_f)
	if err != nil {
		log.Println("exec_queueing() function error", err)
		// とりあえず処理しないでおく
	}

	log.Println("計算結果：", queueing_result.WaitMan, queueing_result.WaitTime, queueing_result.TurnAroundTime)

	kekka = fmt.Sprintf("M/M/%sモデル 待ち %s (人), %s (分), ターンアラウンドタイム %s (分)",
		window,
		queueing_result.WaitMan,
		queueing_result.WaitTime,
		queueing_result.TurnAroundTime)
	// Formへ元データを入力
	form.Window = window
	form.Arrived = arrived
	form.Servises = servises
	// HTMLへ表示
	onDisp = true // 「結果を表示する」へ変更
	c.HTML(http.StatusOK, "index.html", gin.H{
		"form":   form,
		"onDisp": onDisp,
		"kekka":  queueing_result,
		"logs":   kekka,
	})
}

func exec_queueing(ta, ts, win float64) error {
	λ := 1.0 / ta
	u := win * 1.0 / ts

	ro := λ / u // 混雑度

	fmt.Println("結果：", λ, u, math.Pow(ro, win))
	wait_man := math.Pow(ro, win) / (1 - math.Pow(ro, win)) // 待ち人
	wait_time := wait_man * ts
	turn_around_time := wait_time + ts

	queueing_result.WaitMan = strconv.FormatFloat(wait_man, 'f', 1, 64)
	queueing_result.WaitTime = strconv.FormatFloat(wait_time, 'f', 3, 64)
	queueing_result.TurnAroundTime = strconv.FormatFloat(turn_around_time, 'f', 3, 64)

	return nil
}
