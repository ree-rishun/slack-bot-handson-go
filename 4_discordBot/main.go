// 4 Discordっぽい挨拶するやつ
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

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

			// イベントタイプで分岐
			switch event := innerEvent.Data.(type) {
			case *slackevents.MemberJoinedChannelEvent: // チャンネル参加を検知
				// チャンネルに新規参加したとき
				if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText(generateGreeting(event.User), false)); err != nil {
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

// 挨拶生成
func generateGreeting(user string) string {
	rand.Seed(time.Now().Unix())

	greetings := [...]string{
		"<@%s>にご挨拶しな！",
		"いらっしゃい<@%s>ちゃん。ほら、ちゃんとご挨拶して！",
		"<@%s>さん、お会いできて何よりです。",
		"<@%s>がただいま着陸いたしました。",
		"<@%s>が出たぞー！",
		"<@%s>がパーティーに加わりました。",
		"<@%s>さん、ようこそ。",
		"<@%s>がサーバーに飛び乗りました。",
		"<@%s>がサーバーに滑り込みました。",
		"やあ、<@%s>君。ピザ持ってきたよね？",
		"やったー、<@%s>が来たー！",
		"あ！野生の<@%s>が飛び出してきた！",
	}

	return fmt.Sprintf(greetings[rand.Intn(len(greetings))], user)
}
