package controllers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
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
	// ID             uint   `json:"id"`
	WaitMan        string `json:"wait_man"`         // 待ち人(人)
	WaitTime       string `json:"wait_time"`        // 待ち時間(分)
	TurnAroundTime string `json:"turn_around_time"` // ターンアラウンドタイム(分)
}

type QueueingData struct {
	InputData  QueueingTheory
	OutputData QueueingResult
}

var queueing_result QueueingResult // 計算結果を格納

func init() {
	initOutputData()
}

func initOutputData() {
	// queueing_result.ID = 0
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
	var queueings []QueueingData
	var last_que int

	// 初期化
	initOutputData()
	onDisp := false                        // true 表示する, false しない
	const last_queueing_number = 3         // 履歴表示数：最新の何個を表示するかを決める
	const fileName = "queueing_theory.csv" // 履歴を保存するファイル名

	// フォームよりデータを取得
	window := c.PostForm("window")
	arrived := c.PostForm("arrived")
	servises := c.PostForm("servises")

	var form QueueingTheory
	// Bindの代表的２つ　MustBindWith, ShouldBind
	if err := c.ShouldBind(&form); err != nil {
		log.Println("formの取得結果:", form)

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

		// 履歴表示用のcsvを読み込む
		readCSVFile(fileName, &queueings)
		if len(queueings)-last_queueing_number <= 0 {
			last_que = 0 // 指定件数より少ないので、全体を表示する
		} else {
			last_que = len(queueings) - last_queueing_number // 後ろから何個のリストを取得するか
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"form":  form,
			"kekka": queueing_result,
			"logs":  kekka,
			"lasts": queueings[last_que:],
		})
		return
	}
	//

	// Data is OK.

	// 待ち行列を計算する
	err := exec_queueing(arrived_f, servises_f, window_f)
	if err != nil {
		log.Println("exec_queueing() function error", err)
		// とりあえず処理しないでおく
	}

	// CSVで書き出し
	input_param := []string{window, arrived, servises}
	err = toCSVFile(fileName, input_param)
	if err != nil {
		log.Println("toCSVFile() function error", err)

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

	readCSVFile(fileName, &queueings)
	if len(queueings)-last_queueing_number <= 0 {
		last_que = 0 // 指定件数より少ないので、全体を表示する
	} else {
		last_que = len(queueings) - last_queueing_number // 後ろから何個のリストを取得するか
	}
	// log.Println(queueings[last_que:])

	// HTMLへ表示
	onDisp = true // 「結果を表示する」へ変更
	c.HTML(http.StatusOK, "index.html", gin.H{
		"form":   form,
		"onDisp": onDisp,
		"kekka":  queueing_result,
		"logs":   "",
		"lasts":  queueings[last_que:],
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

	// queueing_result.ID += 1
	queueing_result.WaitMan = strconv.FormatFloat(wait_man, 'f', 1, 64)
	queueing_result.WaitTime = strconv.FormatFloat(wait_time, 'f', 3, 64)
	queueing_result.TurnAroundTime = strconv.FormatFloat(turn_around_time, 'f', 3, 64)

	return nil
}

func toCSVFile(fileName string, inp []string) error {
	// 入力チェック
	// 今は入力データは3つを想定
	if len(inp) != 3 {
		msg := "CSV作成入力データエラー"
		log.Println(msg)
		return errors.New(msg)
	}

	fp, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
	// file, err := os.Create(fileName)
	if err != nil {
		log.Println("CSV作成エラー", err)
		return err
	}
	defer fp.Close()
	// fi, _ := file.Stat()
	// leng := fi.Size()
	log.Println("CSVファイル名:", fileName)

	w := csv.NewWriter(fp)
	list := []string{inp[0], inp[1], inp[2], queueing_result.WaitMan, queueing_result.WaitTime, queueing_result.TurnAroundTime}

	err = w.Write(list)
	if err != nil {
		log.Println("CSV書き込みエラー", err)
		return err
	}
	w.Flush()

	return nil
}

func readCSVFile(fileName string, datas *[]QueueingData) error {

	fp, err := os.Open(fileName)
	if err != nil {
		log.Println("CSVファイルオープン込みエラー", err)
		return nil
	}
	defer fp.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(fp).ReadAll()
	if err != nil {
		log.Fatal("CSVファイルエラー", err)
		return err
	}

	var data QueueingData
	for _, line := range lines {
		data.InputData.Window = line[0]
		data.InputData.Arrived = line[1]
		data.InputData.Servises = line[2]
		data.OutputData.WaitMan = line[3]
		data.OutputData.WaitTime = line[4]
		data.OutputData.TurnAroundTime = line[5]
		*datas = append(*datas, data)
	}

	return nil
}
