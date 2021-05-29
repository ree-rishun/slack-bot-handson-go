// 5 並べ替え
package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
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
			case *slackevents.AppMentionEvent: // メンションイベント

				// スペースを区切り文字として配列に格納
				members := strings.Split(event.Text, " ")

				// テキストが送信されていない場合は終了
				if len(members) < 2 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				// タイトル送信
				if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText("<!here>【順番発表】", false)); err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// rand.Seed(time.Now().Unix())

				// 順番カウント
				orderCnt := 0

				// 順番をランダムに並び替え
				for i := 0; len(members)-i-1 > 0; i++ {
					num := rand.Intn(len(members) - i - 1)
					fmt.Println(num)
					fmt.Println(members)
					fmt.Println(len(members))
					// time.Sleep(time.Second * 1)

					// 非メンションを無視
					if strings.Index(members[num], "<@") == 0 {

						// Botへのメンションは無視
						if members[num] != "<@U022K4VR8J2>" {
							// 順位結果格納配列を作成
							if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText(strconv.Itoa(orderCnt+1)+". "+members[num], false)); err != nil {
								log.Println(err)
								w.WriteHeader(http.StatusInternalServerError)
								return
							}
							orderCnt++
						}
					}

					// 入れ替え
					members[num] = members[len(members)-i-1]
				}
			}
		}
	})

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
