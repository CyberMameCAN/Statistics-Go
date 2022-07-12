package main

import (
	"log"
	"net/http"

	"github.com/CyberMameCAN/statistics-go/controllers"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("start server...")
	router := gin.Default()

	router.LoadHTMLGlob("templates/*.html")
	router.Static("/static", "static")

	// ルートなら/static/index.htmlにリダイレクト
	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			// htmlに渡す変数を定義
			// "message": "E=mc^2",
		})
		ctx.Redirect(302, "index.html")
	})

	router.POST("/queueing", controllers.Queueing)

	// サーバを起動
	err := router.Run(":13030")
	if err != nil {
		log.Fatal("サーバ起動に失敗", err)
	}
}
