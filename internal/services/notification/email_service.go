package notification

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"

	"go-finance-advisor/internal/config"
)

// EmailType represents the type of email notification
type EmailType string

const (
	EmailTypeSignal    EmailType = "signal"
	EmailTypeAlert     EmailType = "alert"
	EmailTypePortfolio EmailType = "portfolio"
	EmailTypeSummary   EmailType = "summary"
)

// EmailTemplate holds email template data
type EmailTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// EmailService handles email notifications
type EmailService struct {
	cfg       *config.EmailConfig
	log       *slog.Logger
	templates map[EmailType]*template.Template
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.EmailConfig, log *slog.Logger) *EmailService {
	service := &EmailService{
		cfg:       cfg,
		log:       log,
		templates: make(map[EmailType]*template.Template),
	}

	// Initialize templates
	service.initTemplates()

	return service
}

// initTemplates initializes email templates
func (s *EmailService) initTemplates() {
	// Signal template
	s.templates[EmailTypeSignal] = template.Must(template.New("signal").Parse(`
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background-color: #f9f9f9; }
		.signal-box { background-color: white; padding: 15px; margin: 10px 0; border-left: 4px solid #4CAF50; }
		.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>New Trading Signal</h1>
		</div>
		<div class="content">
			<div class="signal-box">
				<h2>{{.Action}} Signal: {{.Symbol}}</h2>
				<p><strong>Confidence:</strong> {{printf "%.2f" .Confidence}}%</p>
				<p><strong>Asset Type:</strong> {{.AssetType}}</p>
				<p><strong>Entry Price:</strong> ${{printf "%.2f" .EntryPrice}}</p>
				<p><strong>Target Price:</strong> ${{printf "%.2f" .TargetPrice}}</p>
				<p><strong>Stop Loss:</strong> ${{printf "%.2f" .StopLoss}}</p>
				<p><strong>Reasoning:</strong></p>
				<p>{{.Reasoning}}</p>
			</div>
		</div>
		<div class="footer">
			<p>This is an automated notification from GoFinVisor</p>
		</div>
	</div>
</body>
</html>
	`))

	// Alert template
	s.templates[EmailTypeAlert] = template.Must(template.New("alert").Parse(`
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #FF9800; color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background-color: #f9f9f9; }
		.alert-box { background-color: white; padding: 15px; margin: 10px 0; border-left: 4px solid #FF9800; }
		.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Price Alert Triggered</h1>
		</div>
		<div class="content">
			<div class="alert-box">
				<h2>{{.Symbol}}</h2>
				<p><strong>Alert Type:</strong> {{.AlertType}}</p>
				<p><strong>Target Price:</strong> ${{printf "%.2f" .TargetPrice}}</p>
				<p><strong>Current Price:</strong> ${{printf "%.2f" .CurrentPrice}}</p>
				<p><strong>Condition:</strong> {{.Condition}}</p>
				<p><strong>Message:</strong> {{.Message}}</p>
			</div>
		</div>
		<div class="footer">
			<p>This is an automated notification from GoFinVisor</p>
		</div>
	</div>
</body>
</html>
	`))

	// Portfolio summary template
	s.templates[EmailTypeSummary] = template.Must(template.New("summary").Parse(`
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #2196F3; color: white; padding: 20px; text-align: center; }
		.content { padding: 20px; background-color: #f9f9f9; }
		.summary-box { background-color: white; padding: 15px; margin: 10px 0; }
		.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
		.positive { color: #4CAF50; }
		.negative { color: #F44336; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>Portfolio Summary</h1>
		</div>
		<div class="content">
			<div class="summary-box">
				<h2>Your Portfolio Performance</h2>
				<p><strong>Total Value:</strong> ${{printf "%.2f" .TotalValue}}</p>
				<p><strong>Total Invested:</strong> ${{printf "%.2f" .TotalInvested}}</p>
				<p><strong>Unrealized P&L:</strong> <span class="{{if gt .UnrealizedPnL 0}}positive{{else}}negative{{end}}">${{printf "%.2f" .UnrealizedPnL}} ({{printf "%.2f" .ReturnPercent}}%)</span></p>
				<p><strong>Number of Holdings:</strong> {{.HoldingCount}}</p>
			</div>
		</div>
		<div class="footer">
			<p>This is an automated notification from GoFinVisor</p>
		</div>
	</div>
</body>
</html>
	`))
}

// SendEmail sends an email notification
func (s *EmailService) SendEmail(to string, subject string, body string) error {
	if !s.cfg.Enabled {
		s.log.Debug("Email service disabled, skipping email", "to", to)
		return nil
	}

	// Build email message
	msg := s.buildMessage(s.cfg.From, to, subject, body)

	// Send email via SMTP
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	err := smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(msg))
	if err != nil {
		s.log.Error("Failed to send email", "to", to, "error", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.log.Info("Email sent successfully", "to", to, "subject", subject)
	return nil
}

// buildMessage builds the email message with headers
func (s *EmailService) buildMessage(from, to, subject, body string) string {
	msg := fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=UTF-8\r\n"
	msg += "\r\n"
	msg += body
	return msg
}

// SendTemplateEmail sends an email using a template
func (s *EmailService) SendTemplateEmail(to string, emailType EmailType, subject string, data interface{}) error {
	if !s.cfg.Enabled {
		s.log.Debug("Email service disabled, skipping email", "to", to, "type", emailType)
		return nil
	}

	tmpl, ok := s.templates[emailType]
	if !ok {
		return fmt.Errorf("template not found for type: %s", emailType)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return s.SendEmail(to, subject, body.String())
}

// SendSignalNotification sends a signal notification email
func (s *EmailService) SendSignalNotification(to string, data map[string]interface{}) error {
	subject := fmt.Sprintf("New %s Signal: %s", data["Action"], data["Symbol"])
	return s.SendTemplateEmail(to, EmailTypeSignal, subject, data)
}

// SendAlertNotification sends an alert notification email
func (s *EmailService) SendAlertNotification(to string, data map[string]interface{}) error {
	subject := fmt.Sprintf("Price Alert: %s", data["Symbol"])
	return s.SendTemplateEmail(to, EmailTypeAlert, subject, data)
}

// SendPortfolioSummary sends a portfolio summary email
func (s *EmailService) SendPortfolioSummary(to string, data map[string]interface{}) error {
	subject := "Your Daily Portfolio Summary"
	return s.SendTemplateEmail(to, EmailTypeSummary, subject, data)
}

// IsEnabled returns whether the email service is enabled
func (s *EmailService) IsEnabled() bool {
	return s.cfg.Enabled
}
