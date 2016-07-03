## Messengerbot - A bot framework for facebook messenger in golang

Work in progress. contributions welcome.

### Getting started

Its simple to get started, run the following code with your validation token and page access token. Then follow [facebook guide](https://developers.facebook.com/docs/messenger-platform/quickstart) to create a subscription. Now send a message with text "image" to your bot using the facebook page message button. Enjoy creating bots !!!

````
package main

import (
	"net/http"
	"time"
	"github.com/vimukthi-git/messengerbot"
)

func main() {
	w := messengerbot.NewMessengerWebhook("your validation token", "your page access token")
	w.MessageHandler(func(pageId string, s messengerbot.Sender, r messengerbot.Recipient, t time.Time, m messengerbot.IncomingTextMessage) bool {
		switch m.Text {
		case "image":
			w.SendImageMessageByRecipientId(s.Id, "http://messengerdemo.parseapp.com/img/touch.png", "")
			break

		case "button":
			w.SendButtonMessageByRecipientId(s.Id, "VB Super", []messengerbot.Button{
				messengerbot.Button{Type:messengerbot.WEB_URL, Title:"Open Web URL", Url: "https://www.oculus.com/en-us/rift/"},
				messengerbot.Button{Type:messengerbot.POSTBACK, Title:"Call Postback", Payload: "Payload for first bubble"},
			}, "")
			break

		case "generic":
			w.SendGenericMessageByRecipientId(s.Id, []messengerbot.GenericTemplateElement {
				messengerbot.GenericTemplateElement{
					Title: "rift",
					Subtitle: "Next-generation virtual reality",
					ItemUrl: "https://www.oculus.com/en-us/rift/",
					ImageUrl: "http://messengerdemo.parseapp.com/img/rift.png",
					Buttons: []messengerbot.Button {
						messengerbot.Button{Type:messengerbot.WEB_URL, Title:"Open Web URL", Url: "https://www.oculus.com/en-us/rift/"},
						messengerbot.Button{Type:messengerbot.POSTBACK, Title:"Call Postback", Payload: "Payload for first bubble"},
					},
				},
				messengerbot.GenericTemplateElement{
					Title: "touch",
					Subtitle: "Your Hands, Now in VR",
					ItemUrl: "https://www.oculus.com/en-us/touch/",
					ImageUrl: "http://messengerdemo.parseapp.com/img/touch.png",
					Buttons: []messengerbot.Button {
						messengerbot.Button{Type:messengerbot.WEB_URL, Title:"Open Web URL", Url: "https://www.oculus.com/en-us/touch/"},
						messengerbot.Button{Type:messengerbot.POSTBACK, Title:"Call Postback", Payload: "Payload for second bubble"},
					},
				},
			}, "")
			break

		case "receipt":
			w.SendReceiptMessageByRecipientId(
				s.Id, "You", "1232132", "HKD", "paypal",
				"321313213", "https://www.oculus.com/en-us/touch/",
				[]messengerbot.ReceiptTemplateElement{
					messengerbot.ReceiptTemplateElement{
						Title: "1 Oculus VR",
						Subtitle: "VR",
						Quantity: 2,
						Price: 4312.43,
						Currency: "USD",
						ImageUrl: "https://www.oculus.com/en-us/touch/",
					},
				},
				messengerbot.Address{
					Street1: "123/15, sirimangala road",
					Street2: "",
					City: "Makola",
					PostalCode: "11690",
					State: "Colombo",
					// Two-letter country abbreviation
					Country: "LK",
				},
				messengerbot.Summary{
					TotalCost: 321.32,
				},
				[]messengerbot.Adjustment{
					messengerbot.Adjustment{
						Name: "adj",
						Amount: 1,
					},
				},
			"")
			break

		default:
			w.SendTextMessageByRecipientId(s.Id, m.Text, messengerbot.REGULAR)
		}
		return true
	})
	http.HandleFunc("/webhook", w.Handler)
	http.ListenAndServe(":8080", nil)
}

````

### License

Apache 2.0
