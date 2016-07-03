package messengerbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type webhook struct {
	validationToken            string
	pageAccessToken            string
	verifiedCallback           VerifiedCallback
	verificationFailedCallback VerificationFailedCallback
	optinCallback              OptinCallback
	messageCallback            TextMessageCallback
	attachementMessageCallback AttachementMessageCallback
	deliveryCallback           DeliveryCallback
	postbackCallback           PostbackCallback
}

func NewMessengerWebhook(validationToken, pageAccessToken string) *webhook {
	m := new(webhook)
	m.validationToken = validationToken
	m.pageAccessToken = pageAccessToken
	return m
}

func (w *webhook) VerfiedHandler(cb VerifiedCallback) {
	w.verifiedCallback = cb
}

func (w *webhook) VerficationFailedHandler(cb VerificationFailedCallback) {
	w.verificationFailedCallback = cb
}

func (w *webhook) OptinHandler(cb OptinCallback) {
	w.optinCallback = cb
}

func (w *webhook) MessageHandler(cb TextMessageCallback) {
	w.messageCallback = cb
}

func (w *webhook) AttachmentHandler(cb AttachementMessageCallback) {
	w.attachementMessageCallback = cb
}

func (w *webhook) DeliveryHandler(cb DeliveryCallback) {
	w.deliveryCallback = cb
}

func (w *webhook) PostbackHandler(cb PostbackCallback) {
	w.postbackCallback = cb
}

func (w *webhook) Handler(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		hubMode := req.URL.Query().Get("hub.mode")
		hubVerfifyToken := req.URL.Query().Get("hub.verify_token")
		hubChallenge := req.URL.Query().Get("hub.challenge")
		if hubMode == "subscribe" && hubVerfifyToken == w.validationToken {
			log.Println("valid token")
			fmt.Fprintf(res, hubChallenge)
		} else {
			log.Println("invalid token")
			fmt.Fprintf(res, "O")
		}
	} else if req.Method == http.MethodPost {
		log.Println("message received")
		dec := json.NewDecoder(req.Body)
		for {
			var d map[string]interface{}
			if err := dec.Decode(&d); err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			if d["object"] == "page" {
				// process page entries
				for _, pageEntryInterface := range d["entry"].([]interface{}) {
					pageEntry := pageEntryInterface.(map[string]interface{})
					pageId := pageEntry["id"].(string)
					// process events
					for _, messagingEventInterface := range pageEntry["messaging"].([]interface{}) {
						messagingEvent := messagingEventInterface.(map[string]interface{})
						// find sender and receiver
						sender := Sender{messagingEvent["sender"].(map[string]interface{})["id"].(string)}
						recipient := Recipient{Id: messagingEvent["recipient"].(map[string]interface{})["id"].(string)}

						// decide which type of message and handle accordingly
						if optin, ok := messagingEvent["optin"]; ok {
							log.Println("optin : ", optin)
						} else if message, ok := messagingEvent["message"]; ok {
							sentTime := int64(messagingEvent["timestamp"].(float64))
							msg := message.(map[string]interface{})
							if attachments, ok := messagingEvent["attachments"].(map[string]interface{}); ok {
								attachmentPayload := attachments["payload"].(map[string]interface{})
								w.attachementMessageCallback(w, pageId, sender, recipient, time.Unix(sentTime, 0),
									IncomingAttachmentMessage{msg["mid"].(string), msg["seq"].(float64),
										attachments["type"].(string), attachmentPayload["url"].(string)})
							} else {
								w.messageCallback(w, pageId, sender, recipient, time.Unix(sentTime, 0),
									IncomingTextMessage{msg["mid"].(string), msg["seq"].(float64), msg["text"].(string)})
							}
						} else if delivery, ok := messagingEvent["delivery"]; ok {
							log.Println("delivery : ", delivery)
						} else if postback, ok := messagingEvent["postback"]; ok {
							log.Println("postback : ", postback)
						} else {
							log.Println("unknown event : ", messagingEvent)
						}
					}
				}

			}
		}
		fmt.Fprintf(res, "OK")
	}
}

// SendTextMessageByRecipientId send the given message text to the recipient identified by the given
// recipientId
func (w *webhook) SendTextMessageByRecipientId(recipientId, messageText string, notificationType NotificationType) {
	w.callSendApi(MessageEnvelope{Recipient{Id:recipientId}, NewTextMessage(messageText, notificationType)})
}

// SendImageMessageByRecipientId send the image given by the imageUrl to the recipient identified by the given
// recipientId
func (w *webhook) SendImageMessageByRecipientId(recipientId string, imageUrl string, notificationType NotificationType) {
	w.callSendApi(MessageEnvelope{Recipient{Id:recipientId}, NewImageMessage(imageUrl, notificationType)})
}

// SendButtonMessageByRecipientId send the buttons given to the recipient identified by the given
// recipientId
func (w *webhook) SendButtonMessageByRecipientId(recipientId string, text string, buttons []Button, notificationType NotificationType) {
	w.callSendApi(MessageEnvelope{Recipient{Id:recipientId}, NewButtonMessage(text, buttons, notificationType)})
}

// SendGenericMessageByRecipientId send the generic message to the recipient identified by the given
// recipientId
func (w *webhook) SendGenericMessageByRecipientId(recipientId string, elements []GenericTemplateElement, notificationType NotificationType) {
	w.callSendApi(MessageEnvelope{Recipient{Id:recipientId}, NewGenericMessage(elements, notificationType)})
}

// SendReceiptMessageByRecipientId send the receipt message to the recipient identified by the given
// recipientId
func (w *webhook) SendReceiptMessageByRecipientId(recipientId string, recipientName string, orderNumber string,
	currency string, paymentMethod string, timestamp string, orderUrl string, elements []ReceiptTemplateElement,
	shippingAddress Address, paymentSummary Summary, adjustments []Adjustment,
	notificationType NotificationType) {

	w.callSendApi(MessageEnvelope{Recipient{Id:recipientId}, NewReceiptMessage(recipientName, orderNumber,
		currency, paymentMethod,
		timestamp, orderUrl, elements,
		shippingAddress, paymentSummary, adjustments,
		notificationType)})
}

func (w *webhook) callSendApi(data MessageEnvelope) {
	url := "https://graph.facebook.com/v2.6/me/messages?access_token=" + w.pageAccessToken
	jsonStr, e := json.Marshal(data)
	if e != nil {
		log.Fatal("Error in marshalling data")
	}
	log.Println("json : ", string(jsonStr))
	log.Println("url : ", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

