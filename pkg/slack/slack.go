package slack

import (
	"bytes"
	"cercat/config"
	"cercat/pkg/model"
	"encoding/json"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// AttachmentField
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Attachment
type Attachment struct {
	Color    string            `json:"color"`
	Text     string            `json:"text,omitempty"`
	ImageURL string            `json:"image_url,omitempty"`
	Fields   []AttachmentField `json:"fields"`
	// Footer     string                 `json:"footer,omitempty"`
	// FooterIcon string                 `json:"footer_icon,omitempty"`
}

// Payload represents a message to send to Slack
type Payload struct {
	Text        string       `json:"text,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// NewPayload generates a new Slack Payload
func NewPayload(config *config.Configuration, r *model.Result) Payload {
	var attachments []Attachment
	var attachment Attachment
	var fields []AttachmentField
	var field AttachmentField

	field.Title = "Domain"
	field.Value = r.Domain
	field.Short = true
	fields = append(fields, field)

	field.Title = "Issuer"
	field.Value = r.Issuer
	field.Short = true
	fields = append(fields, field)

	if r.IDN != "" {
		field.Title = "IDN"
		field.Value = r.IDN
		field.Short = true
		fields = append(fields, field)
	}

	field.Title = "SAN"
	field.Short = false
	field.Value = strings.Join(r.SAN, ", ")
	fields = append(fields, field)

	field.Title = "Addresses"
	field.Short = false
	field.Value = strings.Join(r.Addresses, ", ")
	fields = append(fields, field)

	attachment.Fields = fields

	attachment.Color = "#ff5400"

	if r.Screenshot != "" {
		attachment.ImageURL = r.Screenshot
	}

	attachments = append(attachments, attachment)

	domain := r.Domain
	if r.IDN != "" {
		domain += " (" + r.IDN + ")"
	}

	return Payload{
		Text:        "A certificate for " + domain + " has been issued",
		Username:    config.SlackUsername,
		IconURL:     config.SlackIconURL,
		Attachments: attachments,
	}
}

// Post posts to Slack a Payload
func (s Payload) Post(config *config.Configuration) {
	body, _ := json.Marshal(s)
	req, _ := http.NewRequest(http.MethodPost, config.SlackWebHookURL, bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	_, err := client.Do(req)
	if err != nil {
		log.Warn("Slack Post error")
	}
}
