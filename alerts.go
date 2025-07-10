package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

type Alerter struct {
	config *Config
}

func NewAlerter(config *Config) *Alerter {
	return &Alerter{
		config: config,
	}
}

func (a *Alerter) SendAlert(result MonitorResult) {
	if a.config.Alerts.Email.Enabled {
		go a.sendEmailAlert(result)
	}
	
	if a.config.Alerts.Slack.Enabled {
		go a.sendSlackAlert(result)
	}
}

func (a *Alerter) sendEmailAlert(result MonitorResult) {
	if !a.config.Alerts.Email.Enabled {
		return
	}
	
	subject := fmt.Sprintf("🚨 ALERT: %s is DOWN", result.Website.Name)
	body := fmt.Sprintf(`
UpTimer Alert - Website Down

Website: %s
URL: %s
Status: DOWN
Status Code: %d
Error: %s
Response Time: %v
Timestamp: %s

Please check the website immediately.

Best regards,
UpTimer Monitoring System
`, result.Website.Name, result.Website.URL, result.StatusCode, result.Error, result.ResponseTime, result.Timestamp.Format(time.RFC3339))
	
	err := a.sendEmail(subject, body)
	if err != nil {
		log.Printf("Failed to send email alert: %v", err)
	} else {
		log.Printf("Email alert sent for %s", result.Website.Name)
	}
}

func (a *Alerter) sendEmail(subject, body string) error {
	cfg := a.config.Alerts.Email
	
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPServer)
	
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", 
		strings.Join(cfg.To, ","), subject, body)
	
	addr := fmt.Sprintf("%s:%d", cfg.SMTPServer, cfg.SMTPPort)
	
	return smtp.SendMail(addr, auth, cfg.From, cfg.To, []byte(msg))
}

func (a *Alerter) sendSlackAlert(result MonitorResult) {
	if !a.config.Alerts.Slack.Enabled {
		return
	}
	
	payload := SlackPayload{
		Channel: a.config.Alerts.Slack.Channel,
		Text:    fmt.Sprintf("🚨 ALERT: %s is DOWN", result.Website.Name),
		Attachments: []SlackAttachment{
			{
				Color: "danger",
				Fields: []SlackField{
					{
						Title: "Website",
						Value: result.Website.Name,
						Short: true,
					},
					{
						Title: "URL",
						Value: result.Website.URL,
						Short: true,
					},
					{
						Title: "Status Code",
						Value: fmt.Sprintf("%d", result.StatusCode),
						Short: true,
					},
					{
						Title: "Response Time",
						Value: result.ResponseTime.String(),
						Short: true,
					},
					{
						Title: "Error",
						Value: result.Error,
						Short: false,
					},
					{
						Title: "Timestamp",
						Value: result.Timestamp.Format(time.RFC3339),
						Short: false,
					},
				},
			},
		},
	}
	
	err := a.sendSlackMessage(payload)
	if err != nil {
		log.Printf("Failed to send Slack alert: %v", err)
	} else {
		log.Printf("Slack alert sent for %s", result.Website.Name)
	}
}

func (a *Alerter) sendSlackMessage(payload SlackPayload) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(a.config.Alerts.Slack.WebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}
	
	return nil
}

type SlackPayload struct {
	Channel     string            `json:"channel"`
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments"`
}

type SlackAttachment struct {
	Color  string       `json:"color"`
	Fields []SlackField `json:"fields"`
}

type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}