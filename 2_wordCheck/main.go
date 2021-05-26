// 2 禁止用語チェッカー
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func main() {
	// .envの読み取り
	godotenv.Load(fmt.Sprintf("../%s.env", os.Getenv("GO_ENV")))

	// SlackClientの構築
	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	// ルートにアクセスがあった時の処理
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch eventsAPIEvent.Type {
		case slackevents.URLVerification: // URL検証の場合の処理
			var res *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte(res.Challenge)); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case slackevents.CallbackEvent: // コールバックイベントの場合の処理
			innerEvent := eventsAPIEvent.InnerEvent

			// イベントタイプで分岐
			switch event := innerEvent.Data.(type) {
			case *slackevents.MessageEvent: // メッセージイベント
				if strings.Index(event.Text, "fuck you") != -1 {
					// 送信元のユーザIDを取得
					user := event.User

					// 送信元ユーザに注意
					if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("<@"+user+"> が禁止用語を発言しました。", false)); err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			}
		}
	})

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
