package messengerbot

import (
	"testing"
	"net/http"
	"time"
)

func TestWebhook(t *testing.T) {
	config := getTestConfig();
	w := NewMessengerWebhook(config.ValidationToken, config.PageAccessToken)
	w.MessageHandler(func(pageId string, s Sender, r Recipient, t time.Time, m IncomingTextMessage) bool {
		switch m.Text {
		case "image":
			w.SendImageMessageByRecipientId(s.Id, "http://messengerdemo.parseapp.com/img/touch.png", "")
			break

		case "button":
			w.SendButtonMessageByRecipientId(s.Id, "VB Super", []Button{
				Button{Type:WEB_URL, Title:"Open Web URL", Url: "https://www.oculus.com/en-us/rift/"},
				Button{Type:POSTBACK, Title:"Call Postback", Payload: "Payload for first bubble"},
			}, "")
			break

		case "generic":
			w.SendGenericMessageByRecipientId(s.Id, []GenericTemplateElement {
				GenericTemplateElement{
					Title: "rift",
					Subtitle: "Next-generation virtual reality",
					ItemUrl: "https://www.oculus.com/en-us/rift/",
					ImageUrl: "http://messengerdemo.parseapp.com/img/rift.png",
					Buttons: []Button {
						Button{Type:WEB_URL, Title:"Open Web URL", Url: "https://www.oculus.com/en-us/rift/"},
						Button{Type:POSTBACK, Title:"Call Postback", Payload: "Payload for first bubble"},
					},
				},
				GenericTemplateElement{
					Title: "touch",
					Subtitle: "Your Hands, Now in VR",
					ItemUrl: "https://www.oculus.com/en-us/touch/",
					ImageUrl: "http://messengerdemo.parseapp.com/img/touch.png",
					Buttons: []Button {
						Button{Type:WEB_URL, Title:"Open Web URL", Url: "https://www.oculus.com/en-us/touch/"},
						Button{Type:POSTBACK, Title:"Call Postback", Payload: "Payload for second bubble"},
					},
				},
			}, "")
			break

		case "receipt":
			w.SendReceiptMessageByRecipientId(
				s.Id, "You", "1232132", "HKD", "paypal",
				"321313213", "https://www.oculus.com/en-us/touch/",
				[]ReceiptTemplateElement{
					ReceiptTemplateElement{
						Title: "1 Oculus VR",
						Subtitle: "VR",
						Quantity: 2,
						Price: 4312.43,
						Currency: "USD",
						ImageUrl: "https://www.oculus.com/en-us/touch/",
					},
				},
				Address{
					Street1: "123/15, sirimangala road",
					Street2: "",
					City: "Makola",
					PostalCode: "11690",
					State: "Colombo",
					// Two-letter country abbreviation
					Country: "LK",
				},
				Summary{
					TotalCost: 321.32,
				},
				[]Adjustment{
					Adjustment{
						Name: "adj",
						Amount: 1,
					},
				},
			"")
			break

		default:
			w.SendTextMessageByRecipientId(s.Id, m.Text, REGULAR)
		}
		return true
	})
	http.HandleFunc("/webhook", w.Handler)
	http.ListenAndServe(":8080", nil)
}