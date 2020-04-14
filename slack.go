package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type slackAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackAttachment struct {
	Color  string                 `json:"color"`
	Text   string                 `json:"text,omitempty"`
	Fields []slackAttachmentField `json:"fields"`
	// Footer     string                 `json:"footer,omitempty"`
	// FooterIcon string                 `json:"footer_icon,omitempty"`
}

type slackPayload struct {
	Text        string            `json:"text,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconURL     string            `json:"icon_url,omitempty"`
	Attachments []slackAttachment `json:"attachments,omitempty"`
}

func newSlackPayload(r result) slackPayload {
	var attachments []slackAttachment
	var attachment slackAttachment
	var fields []slackAttachmentField
	var field slackAttachmentField

	field.Title = "Domain"
	field.Value = r.Domain
	field.Short = true
	fields = append(fields, field)
	field.Title = "Issuer"
	field.Value = r.Issuer
	field.Short = true
	fields = append(fields, field)

	var s string
	for _, i := range r.SAN {
		s += i + ", "
	}
	field.Title = "SAN"
	field.Short = false
	field.Value = s[:len(s)-2]
	fields = append(fields, field)

	s = ""
	for _, i := range r.Addresses {
		s += i + ", "
	}
	field.Title = "Addresses"
	field.Short = false
	field.Value = s[:len(s)-2]
	fields = append(fields, field)

	attachment.Fields = fields

	attachment.Color = "#ff5400"

	attachments = append(attachments, attachment)

	return slackPayload{
		Text:        "A certificate for *" + r.Domain + "* has been issued",
		Username:    config.SlackUsername,
		IconURL:     config.SlackIconURL,
		Attachments: attachments}
}

func (s slackPayload) Post() {
	body, _ := json.Marshal(s)
	req, _ := http.NewRequest(http.MethodPost, config.SlackWebHookURL, bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	_, err := client.Do(req)
	if err != nil {
		log.Println("[ERROR] : Slack Post error")
	}
}
