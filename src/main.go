package main

//
// メイン機能
//   待ち行列の計算をするアプリ
//
// 2023.01.16
//   JWT認証をする機能を追加
//   JWTで作ったトークンをcookieにセットして認証に使う
//

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/CyberMameCAN/statistics-go/controllers"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var mySigningKey []uint8

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(".env Not Found")
	}
	mySigningKey = []byte(os.Getenv("SIGNINGKEY"))
	// fmt.Printf("%T, mySigningKey)  // 型を表示する
}

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*.html")
	router.Static("/static", "static")

	// リバースプロキシの設定も出来る
	auth := router.Group("/auth")
	{
		// With Authorization
		authorized := auth.Group("/")
		{
			authorized.GET("/", verifyAuthorized(myPage))
			authorized.GET("/private", verifyAuthorized(secretPage))
			authorized.GET("/admin", verifyAuthorized(myPage))
		}
	}

	// Without Authorization
	router.GET("/healthcheck", ReverseProxyHandler)
	// ログイン処理をする
	router.GET("/login", LoginPage)

	// ルートなら/static/index.htmlにリダイレクト
	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", gin.H{
			// htmlに渡す変数を定義
			// "message": "E=mc^2",
		})
		ctx.Redirect(302, "index.html")
	})

	router.POST("/queueing", controllers.Queueing)

	// cookieセットテスト
	router.GET("/cookie", func(ctx *gin.Context) {
		cookie, err := ctx.Cookie("gin_cookie")
		if err != nil {
			cookie = "NotSet"
			ctx.SetCookie("gin_cookie", "test", 300, "/", "localhost", false, true)
		}
		fmt.Printf("Cookie value: %s \n", cookie)
	})

	// PreJWt()
	// AfterJWt()

	// サーバを起動
	err := router.Run(":13030")
	if err != nil {
		log.Fatal("サーバ起動に失敗", err)
	}
}

// JWT トークンの検証 GinのHandlerFuncを返す？
// トークンはcookieより取得するように変更
func verifyAuthorized(endpoint gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, _ := ctx.Cookie("gin_cookie")

		// if ctx.Request.Header["Authorization"] != nil { // トークンを検証する
		if cookie != "" { // トークンを検証する
			// token, err := jwt.Parse(ctx.Request.Header["Authorization"][0], func(token *jwt.Token) (interface{}, error) {
			token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // 1行での書き方 okでなかったら return 処理へ
					return nil, fmt.Errorf("there was an error")
				}
				return mySigningKey, nil
			})

			if err != nil {
				ctx.String(200, err.Error())
			}

			if token.Valid {
				endpoint(ctx)
			}

		} else {
			ctx.String(200, "Not Authorized")
		}
	}
}

func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256) // 新しいトークンを作成

	claims := token.Claims.(jwt.MapClaims) // claimsを変更する
	claims["authorized"] = true
	claims["user"] = "mamecan"                             // (※)ここをユーザidにいずれ変更する
	claims["exp"] = time.Now().Add(time.Minute * 5).Unix() // 有効期限を指定
	// claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString(mySigningKey) // 秘密鍵を使用して文字列に署名する
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ReverseProxyHandler(ctx *gin.Context) {
	// ctx.String(200, "Super Secret Information")
	ctx.HTML(http.StatusOK, "message.html", gin.H{
		"title":   "タイトル サンプル",
		"message": "E=mc^2",
	})
}

func LoginPage(ctx *gin.Context) {
	tokenString, err := GenerateJWT()
	if err != nil {
		fmt.Println("Error generating token string")
	}

	ctx.SetCookie("gin_cookie", tokenString, 180, "/", "localhost", false, true)
	// fmt.Println(tokenString)

	ctx.HTML(http.StatusOK, "message.html", gin.H{
		"title":   "ログイン結果",
		"message": "Maybe OK...",
	})
}

func secretPage(ctx *gin.Context) {
	// ctx.String(200, "Super Secret Information")
	ctx.HTML(http.StatusOK, "message.html", gin.H{
		// htmlに渡す変数を定義
		"title":   "報告",
		"message": "認証完了しました",
	})
	// ctx.Redirect(302, "message.html")
}

func myPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "message.html", gin.H{
		"title":   "マイページのルート",
		"message": "マイページのルートにアクセスしました。",
	})
}

func PreJWt() {
	// Claimsオブジェクトの作成
	claims := jwt.MapClaims{
		"user_id": 12345678,
		"exp":     time.Now().Add(time.Minute * 3).Unix(),
	}

	// ヘッダーとペイロードの生成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	fmt.Printf("Header: %#v\n", token.Header) // Header: map[string]interface {}{"alg":"HS256", "typ":"JWT"}
	fmt.Printf("Claims: %#v\n", token.Claims) // CClaims: jwt.MapClaims{"exp":1634051243, "user_id":12345678}

	// トークンに署名を付与
	tokenString, _ := token.SignedString([]byte("SECRET_KEY"))
	fmt.Println("tokenString:", tokenString) // tokenString: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MzQwNTEzNjMsInVzZXJfaWQiOjEyMzQ1Njc4fQ.OooYrharapD5X2LV5UUWBOkEqH57wDfMd5ibkIpJHYM
}

func AfterJWt() {
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NzM5MTcyOTQsInVzZXJfaWQiOjg4OTA4Nn0.ien3uOCD8BIoUyByWbVHEzxlP2m87snjF06pd3pqcms"

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte("SECRET_KEY"), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Printf("user_id: %v\n", int64(claims["user_id"].(float64)))
		fmt.Printf("exp: %v\n", int64(claims["exp"].(float64)))
	} else {
		fmt.Println(err)
	}
}
