package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

var icon_url = "https://github.com/Caiyeon/goldfish/raw/master/frontend/client/assets/logo_small.png"

func PostMessageWebhook(channel, main_text, attachment_text, webhook string) (err error) {
	payload, err := json.Marshal(
		map[string]interface{}{
			"channel":  channel,
			"username": "Goldfish Vault UI",
			"icon_url": icon_url,
			"text":     main_text,
			"attachments": []interface{}{
				map[string]interface{}{
					"mrkdwn_in":   []string{"text"},
					"text":        attachment_text,
					"footer":      "<https://github.com/Caiyeon/goldfish|Goldfish Vault UI>",
					"footer_icon": icon_url,
					"ts":          time.Now().Unix(),
				},
			},
		},
	)
	if err == nil {
		_, err = http.Post(webhook, "application/json", bytes.NewReader(payload))
	}
	return
}
