// Package p contains an HTTP Cloud Function.
package p

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/logging"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("HelloWorld", HelloWorld)
}

// local,gcp 環境むけにloggerを作る。
func CreateLogger() *log.Logger {

	_, isExist := os.LookupEnv("FUNCTION_SIGNATURE_TYPE")

	var logger *log.Logger

	// gcp環境かどうかの確認。
	if isExist {

		ctx := context.Background()

		// Sets your Google Cloud Platform project ID.
		client, err := logging.NewClient(ctx, os.Getenv("GCP_PROJECT"))
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		logName := "my-log"

		logger = client.Logger(logName).StandardLogger(logging.Info)
	} else {

		logger = log.New(os.Stdout, "", log.LstdFlags|log.LUTC)
	}

	return logger
}

// 適切なoriginを設定する。
func CreateAccessControlAllowOrigin() string {

	accessAllowOriginList := []string{"https://katsutoshiotogawa.github.io"}

	// テスト環境ならlocalhostからも見れるようにしたいので、https://127.0.0.1を有効にしておく。

	_, isExist := os.LookupEnv("LOCAL_ENV")
	if isExist {
		accessAllowOriginList = append(accessAllowOriginList, "https://127.0.0.1")
	}

	return strings.Join(accessAllowOriginList, " ")
}

// HelloWorld prints the JSON encoded "message" field in the body
// of the request or "Hello, World!" if there isn't one.
func HelloWorld(w http.ResponseWriter, r *http.Request) {

	logger := CreateLogger()
	api_url := "https://opendata.resas-portal.go.jp/api/v1/prefectures"
	req, err := http.NewRequest("GET", api_url, nil)

	if err != nil {

		logger.Println("can't create request")
		// bad requestの原因を書く。
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	req.Header.Set("X-API-KEY", os.Getenv("PREFECTURE_API_KEY"))
	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {

		logger.Println("can't create response")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logger.Println("can't write response body")
		// reponse
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	origin := CreateAccessControlAllowOrigin()
	w.Header().Add("Access-Control-Allow-Origin", origin)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// jsonをクライアントに返す。
	fmt.Fprint(w, string(byteArray))
}
