package messengerbot

import (
	"time"
)

type IncomingTextMessage struct {
	Mid  string  `json:"mid,omitempty"`
	Seq  float64 `json:"seq,omitempty"`
	Text string  `json:"text"`
	QuickReply *QuickReply `json:"quick_reply"`
}

type IncomingAttachmentMessage struct {
	Mid  string  `json:"mid,omitempty"`
	Seq  float64 `json:"seq,omitempty"`
	AttachmentType string
	AttachmentUrl string
}

type EventDelivery struct {
	Mid  string  `json:"mid,omitempty"`
	Watermark  float64 `json:"watermark,omitempty"`
	Seq  float64 `json:"seq,omitempty"`
}

type EventPostback  struct {
	Payload  string  `json:"payload,omitempty"`
}


type VerifiedCallback func() string

type VerificationFailedCallback func() string

type OptinCallback func() string

type TextMessageCallback func(string, Sender, Recipient, time.Time, IncomingTextMessage) bool

type AttachementMessageCallback func(string, Sender, Recipient, time.Time, IncomingAttachmentMessage) bool

type DeliveryCallback func(string, Sender, Recipient, EventDelivery) bool

type PostbackCallback func(string, Sender, Recipient, time.Time, EventPostback) bool

// send api
// https://developers.facebook.com/docs/messenger-platform/send-api-reference

type SenderActionType string

const (
	TYPING_ON SenderActionType = "typing_on"
	TYPING_OFF SenderActionType = "typing_off"
	MARK_SEEN SenderActionType = "mark_seen"
)

type NotificationType string

const (
	REGULAR NotificationType = "REGULAR"
	SILENT_PUSH NotificationType = "SILENT_PUSH"
	NO_PUSH NotificationType = "NO_PUSH"
)

type PayloadType string

const (
	IMAGE PayloadType = "image"
	TEMPLATE PayloadType = "template"
)

type TemplateType string

const (
	GENERIC TemplateType = "generic"
	BUTTON TemplateType = "button"
	RECEIPT TemplateType = "receipt"
)

type QuickReplyContentType string

const (
	TEXT QuickReplyContentType = "text"
)

type ButtonType string

const (
	WEB_URL ButtonType = "web_url"
	POSTBACK ButtonType = "postback"
)

type Message struct {
	Text string  `json:"text,omitempty"`
	Attachment *Attachment  `json:"attachment,omitempty"`
	QuickReplies []QuickReply `json:"quick_replies,omitempty"`
}

type QuickReply struct {
	ContentType  QuickReplyContentType  `json:"content_type,omitempty"`
	Title  string  `json:"title,omitempty"`
	Payload string `json:"payload,omitempty"`
}

func NewTextMessage(text string, quickReplies []QuickReply) *Message {
	m := new(Message)
	m.Text = text
	m.Attachment = nil
	m.QuickReplies = quickReplies
	return m
}

type Attachment struct {
	Type  PayloadType  `json:"type"`
	Payload AttachmentPayload `json:"payload"`
}

type AttachmentPayload interface {
	AttachmentPayloadType() PayloadType
}

type ImagePayload struct {
	Url string `json:"url"`
}

func (a ImagePayload) AttachmentPayloadType() PayloadType {
	return IMAGE
}

func NewImageMessage(url string, quickReplies []QuickReply) *Message {
	m := new(Message)
	i := ImagePayload{url}
	a := &Attachment{i.AttachmentPayloadType(), i}
	m.Attachment = a
	m.QuickReplies = quickReplies
	return m
}

type GenericTemplate struct {
	TemplateType  TemplateType  `json:"template_type"`
	Elements []GenericTemplateElement `json:"elements"`
}

func (a GenericTemplate) AttachmentPayloadType() PayloadType {
	return TEMPLATE
}

type GenericTemplateElement struct {
	Title string `json:"title"`
	ItemUrl string `json:"item_url,omitempty"`
	ImageUrl string `json:"image_url,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Buttons []Button `json:"buttons,omitempty"`
}

func NewGenericMessage(elements []GenericTemplateElement, quickReplies []QuickReply) *Message {
	m := new(Message)
	i := GenericTemplate{GENERIC, elements}
	a := &Attachment{i.AttachmentPayloadType(), i}
	m.Attachment = a
	m.QuickReplies = quickReplies
	return m
}

type ButtonTemplate struct {
	TemplateType  TemplateType  `json:"template_type"`
	Text string `json:"text"`
	Buttons []Button `json:"buttons,omitempty"`
}

func NewButtonMessage(text string, buttons []Button, quickReplies []QuickReply) *Message {
	m := new(Message)
	i := ButtonTemplate{BUTTON, text, buttons}
	a := &Attachment{i.AttachmentPayloadType(), i}
	m.Attachment = a
	m.QuickReplies = quickReplies
	return m
}

func (a ButtonTemplate) AttachmentPayloadType() PayloadType {
	return TEMPLATE
}

type Button struct {
	Type ButtonType `json:"type"`
	Title string `json:"title"`
	Url string `json:"url,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type ReceiptTemplate struct {
	TemplateType  TemplateType  `json:"template_type"`
	RecipientName string `json:"recipient_name"`
	OrderNumber string `json:"order_number"`
	Currency string `json:"currency"`
	PaymentMethod string `json:"payment_method"`
	Timestamp string `json:"timestamp,omitempty"`
	OrderUrl string `json:"order_url,omitempty"`
	Elements []ReceiptTemplateElement `json:"elements"`
	ShippingAddress Address `json:"address,omitempty"`
	PaymentSummary Summary `json:"summary"`
	Adjustments []Adjustment `json:"adjustments,omitempty"`
}

func NewReceiptMessage(recipientName string, orderNumber string, currency string, paymentMethod string,
		timestamp string, orderUrl string, elements []ReceiptTemplateElement,
		shippingAddress Address, paymentSummary Summary, adjustments []Adjustment,
		quickReplies []QuickReply) *Message {
	m := new(Message)
	i := ReceiptTemplate{
		RECEIPT,
		recipientName,
		orderNumber,
		currency,
		paymentMethod,
		timestamp,
		orderUrl,
		elements,
		shippingAddress,
		paymentSummary,
		adjustments,
	}
	a := &Attachment{i.AttachmentPayloadType(), i}
	m.Attachment = a
	m.QuickReplies = quickReplies
	return m
}

func (a ReceiptTemplate) AttachmentPayloadType() PayloadType {
	return TEMPLATE
}

type ReceiptTemplateElement struct {
	Title string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	Quantity int64 `json:"quantity,omitempty"`
	Price float64 `json:"price,omitempty"`
	Currency string `json:"currency,omitempty"`
	ImageUrl string `json:"image_url,omitempty"`
}

type Address struct {
	Street1 string `json:"street_1"`
	Street2 string `json:"street_2,omitempty"`
	City string `json:"city"`
	PostalCode string `json:"postal_code"`
	State string `json:"state"`
	// Two-letter country abbreviation
	Country string `json:"country"`
}

type Summary struct {
	Subtotal float64 `json:"subtotal,omitempty"`
	ShippingCost float64 `json:"shipping_cost,omitempty"`
	TotalTax float64 `json:"total_tax,omitempty"`
	TotalCost float64 `json:"total_cost"`
}

type Adjustment struct {
	Name string `json:"name,omitempty"`
	Amount float64 `json:"amount,omitempty"`
}

type Recipient struct {
	Id string `json:"id"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type Sender struct {
	Id string `json:"id"`
}

type MessageEnvelope struct {
	Recipient Recipient `json:"recipient"`
	Message   *Message   `json:"message"`
	SenderAction SenderActionType `json:"sender_action"`
	NotificationType NotificationType `json:"notification_type,omitempty"`
}
