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
	attachmentMessageCallback  AttachementMessageCallback
	deliveryCallback           DeliveryCallback
	postbackCallback           PostbackCallback
}

func NewMessengerWebhook(validationToken, pageAccessToken string) *webhook {
	m := new(webhook)
	m.validationToken = validationToken
	m.pageAccessToken = pageAccessToken
	m.verifiedCallback = func() string {log.Println("Default verfied callback called"); return ""}
	m.verificationFailedCallback = func() string {log.Println("Default verfication failed callback called"); return ""}
	m.optinCallback = func() string {log.Println("Default optin callback called"); return ""}
	m.messageCallback = func(id string, s Sender, r Recipient,
		t time.Time, i IncomingTextMessage) bool {log.Println("Default text message callback called"); return true}
	m.attachmentMessageCallback = func(id string, s Sender, r Recipient,
		t time.Time, i IncomingAttachmentMessage) bool {
		log.Println("Default attachment message callback called"); return true}
	m.deliveryCallback = func(id string, s Sender, r Recipient,
		e EventDelivery) bool {log.Println("Default delivery callback called"); return true}
	m.postbackCallback = func(id string, s Sender, r Recipient,
		t time.Time, e EventPostback) bool {log.Println("Default postback callback called"); return true}
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
	w.attachmentMessageCallback = cb
}

func (w *webhook) DeliveryHandler(cb DeliveryCallback) {
	w.deliveryCallback = cb
}

func (w *webhook) PostbackHandler(cb PostbackCallback) {
	w.postbackCallback = cb
}

func (w *webhook) Handler(res http.ResponseWriter, req *http.Request) {

	////////////////////////////////////////////////////////
	// TODO REFACTOR THIS SHIT
	////////////////////////////////////////////////////////
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
							if attachments, ok := msg["attachments"].([]interface{}); ok {
								log.Println(attachments)
								for _, attachment := range attachments {
									stop := false
									attachmentMap, ok_attachmentMap := attachment.(map[string]interface{})
									if !ok_attachmentMap {
										stop = true
										log.Println("warning: cannot attachment to map[string]interface{}")
									}

									attachmentPayload, ok_attachmentPayload := attachmentMap["payload"].(map[string]interface{})
									if !ok_attachmentPayload {
										stop = true
										log.Println("warning: cannot cast attachmentMap[\"payload\"] to map[string]interface{}")
									}

									str_mid, ok_mid := msg["mid"].(string)
									if !ok_mid {
										stop = true
										log.Println("warning: cannot cast msg[\"mid\"] to string", msg["mid"])
									}

									float_seq, ok_seq := msg["seq"].(float64);
									if !ok_seq {
										stop = true
										log.Println("warning: cannot cast msg[\"seq\"] to float")
									}

									str_type, ok_type := attachmentMap["type"].(string)
									if !ok_type {
										stop = true
										log.Println("warning: cannot cast attachmentMap[\"type\"] to string")
									}

									str_url, ok_url := attachmentPayload["url"].(string);
									if !ok_url {
										stop = true
										log.Println("warning: cannot cast attachmentPayload[\"url\"].(string)} to string")
									}

									if !stop {
										w.attachmentMessageCallback(pageId, sender, recipient, time.Unix(sentTime, 0),
										IncomingAttachmentMessage{str_mid, float_seq, str_type, str_url})
									} else {
										log.Println("warning: attachmentMessageCallback stopped due to casting errors")
									}
								}
							} else {
								stop := false

								str_mid, ok_mid := msg["mid"].(string)
								if !ok_mid {
									stop = true
									log.Println("warning: cannot cast msg[\"mid\"] to string")
								}

								str_text, ok_text := msg["text"].(string)
								if !ok_text {
									stop = true
									log.Println("warning: cannot cast msg[\"text\"] to string")
								}

								float_seq, ok_seq := msg["seq"].(float64);
								if !ok_seq {
									stop = true
									log.Println("warning: cannot cast msg[\"seq\"] to float")
								}

								if !stop {
									w.messageCallback(pageId, sender, recipient, time.Unix(sentTime, 0),
									IncomingTextMessage{str_mid, float_seq, str_text})
								} else {
									log.Println("warning: messageCallback stopped due to casting errors")
								}
							}
						} else if delivery, ok := messagingEvent["delivery"]; ok {
							del := delivery.(map[string]interface{})
							stop := false
							float_seq, ok_seq := del["seq"].(float64);
							if !ok_seq {
								stop = true
								log.Println("warning: cannot cast del[\"seq\"] to float")
							}

							float_wmrk, ok_wmrk := del["watermark"].(float64);
							if !ok_wmrk {
								stop = true
								log.Println("warning: cannot cast del[\"watermark\"] to float")
							}

							if !stop {
								mids, ok_mids := del["mids"].([]interface{});

								if !ok_mids {
									log.Println("warning: deliveryCallback stopped, del[\"mids\"] is not an array")
								} else {
									for _, mid := range mids {
										nested := false

										str_mid, ok_mid := mid.(string)
										if !ok_mid {
											nested = true
											log.Println("warning: cannot cast object of mids to string")
										}

										if !nested {
											w.deliveryCallback(pageId, sender, recipient,
											EventDelivery{str_mid, float_wmrk, float_seq})
										} else {
											log.Println("warning: deliveryCallback stopped due to casting errors")
										}
									}
								}
							} else {
								log.Println("warning: all deliveryCallback's stopped due to casting errors")
							}
						} else if postback, ok := messagingEvent["postback"]; ok {
							sentTime := int64(messagingEvent["timestamp"].(float64))
							pos := postback.(map[string]interface{})

							stop := false

							str_payload, ok_payload := pos["payload"].(string)
							if !ok_payload {
								stop = true
								log.Println("warning: cannot cast pos[\"payload\"] to string")
							}

							if !stop {
								w.postbackCallback(pageId, sender, recipient, time.Unix(sentTime, 0),
								EventPostback{str_payload})
							} else {
								log.Println("warning: postbackCallback stopped due to casting errors")
							}
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

// SendSenderActionByRecipientId send the given message text to the recipient identified by the given
// recipientId
func (w *webhook) SendSenderActionByRecipientId(recipientId string, senderAction SenderActionType) {
	w.callSendApi(MessageEnvelope{Recipient{Id:recipientId}, nil,
		senderAction, ""})
}

// SendTextMessageByRecipientId send the given message text to the recipient identified by the given
// recipientId
func (w *webhook) SendTextMessageByRecipientId(recipientId, messageText string,
	quickReplies []QuickReply, notificationType NotificationType) {
	w.callSendApi(MessageEnvelope{Recipient{Id:recipientId}, NewTextMessage(messageText, quickReplies),
		"", notificationType})
}

// SendImageMessageByRecipientId send the image given by the imageUrl to the recipient identified by the given
// recipientId
func (w *webhook) SendImageMessageByRecipientId(recipientId, imageUrl string, quickReplies []QuickReply,
	notificationType NotificationType) {
	w.callSendApi(MessageEnvelope{Recipient{Id:recipientId}, NewImageMessage(imageUrl, quickReplies), "", notificationType})
}

// SendButtonMessageByRecipientId send the buttons given to the recipient identified by the given
// recipientId
func (w *webhook) SendButtonMessageByRecipientId(recipientId, text string, buttons []Button,
	quickReplies []QuickReply, notificationType NotificationType) {
	w.callSendApi(MessageEnvelope{Recipient{Id:recipientId}, NewButtonMessage(text, buttons, quickReplies),
		"", notificationType})
}

// SendGenericMessageByRecipientId send the generic message to the recipient identified by the given
// recipientId
func (w *webhook) SendGenericMessageByRecipientId(recipientId string, elements []GenericTemplateElement,
	quickReplies []QuickReply, notificationType NotificationType) {
	w.callSendApi(MessageEnvelope{Recipient{Id:recipientId}, NewGenericMessage(elements, quickReplies),
		"", notificationType})
}

// SendReceiptMessageByRecipientId send the receipt message to the recipient identified by the given
// recipientId
func (w *webhook) SendReceiptMessageByRecipientId(recipientId, recipientName, orderNumber,
	currency, paymentMethod, timestamp, orderUrl string, elements []ReceiptTemplateElement,
	shippingAddress Address, paymentSummary Summary, adjustments []Adjustment, quickReplies []QuickReply,
	notificationType NotificationType) {

	w.callSendApi(MessageEnvelope{
		Recipient{Id:recipientId},
		NewReceiptMessage(
			recipientName, orderNumber,
			currency, paymentMethod,
			timestamp, orderUrl, elements,
			shippingAddress, paymentSummary, adjustments,
			quickReplies,
		),
		"",
		notificationType,
	})
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
