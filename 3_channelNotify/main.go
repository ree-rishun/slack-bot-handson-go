// 3 チャンネル追加通知
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
		// リクエスト内容を取得
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// イベント内容を取得
		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// イベント内容によって処理を分岐
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

			notifyChannelID := "C022YFGH11S"

			// イベントタイプで分岐
			switch event := innerEvent.Data.(type) {
			case *slackevents.ChannelCreatedEvent: // チャンネル作成を検知
				// チャンネル作成通知を送信
				if _, _, err := api.PostMessage(notifyChannelID, slack.MsgOptionText("<#"+event.Channel.ID+"> を <@"+event.Channel.Creator+"> が作成しました。", false)); err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
	})

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
