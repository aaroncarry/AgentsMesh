package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"

	"github.com/resend/resend-go/v2"

	"net/smtp"
	"strings"
	"time"
)

// Service defines the email service interface
type Service interface {
	// SendVerificationEmail sends an email verification link
	SendVerificationEmail(ctx context.Context, to, token string) error

	// SendPasswordResetEmail sends a password reset link
	SendPasswordResetEmail(ctx context.Context, to, token string) error

	// SendOrgInvitationEmail sends an organization invitation
	SendOrgInvitationEmail(ctx context.Context, to, orgName, inviterName, token string) error
}

// RenewalReminderSender is an optional interface for sending renewal reminders
type RenewalReminderSender interface {
	// SendRenewalReminder sends a subscription renewal reminder email
	SendRenewalReminder(ctx context.Context, to, orgName, planName string, expiryDate time.Time, daysRemaining int, orgSlug string) error
}

// Config holds email service configuration
type Config struct {
	Provider    string // "resend", "smtp", or "console" (for development)
	ResendKey   string
	FromAddress string // e.g., "AgentsMesh <noreply@agentsmesh.dev>"
	BaseURL     string // Frontend base URL for links, e.g., "https://agentsmesh.dev"

	// SMTP configuration (used when Provider == "smtp")
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

// NewService creates a new email service based on configuration
func NewService(cfg Config) Service {
	switch cfg.Provider {
	case "smtp":
		// Use SMTP if configured
		if cfg.SMTPHost != "" {
			return &SMTPService{
				host:     cfg.SMTPHost,
				port:     cfg.SMTPPort,
				username: cfg.SMTPUsername,
				password: cfg.SMTPPassword,
				from:     cfg.SMTPFrom,
				baseURL:  cfg.BaseURL,
			}
		}
		// Fall back to console if SMTP not configured
		return &ConsoleService{
			baseURL: cfg.BaseURL,
		}
	case "resend":
		if cfg.ResendKey != "" {
			return &ResendService{
				client:      resend.NewClient(cfg.ResendKey),
				fromAddress: cfg.FromAddress,
				baseURL:     cfg.BaseURL,
			}
		}
		// Fall back to console if Resend not configured
		return &ConsoleService{
			baseURL: cfg.BaseURL,
		}
	default:
		// Default to console
		return &ConsoleService{
			baseURL: cfg.BaseURL,
		}
	}
}

// ResendService implements email sending via Resend
type ResendService struct {
	client      *resend.Client
	fromAddress string
	baseURL     string
}

// SendVerificationEmail sends email verification via Resend
func (s *ResendService) SendVerificationEmail(ctx context.Context, to, token string) error {
	verifyURL := fmt.Sprintf("%s/verify-email/callback?token=%s", s.baseURL, token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Verify your email</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #333;">Welcome to AgentsMesh!</h1>
    <p style="color: #666; font-size: 16px;">Please verify your email address by clicking the button below:</p>
    <p style="margin: 30px 0;">
        <a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 500;">Verify Email</a>
    </p>
    <p style="color: #999; font-size: 14px;">Or copy this link: <a href="%s" style="color: #0070f3;">%s</a></p>
    <p style="color: #999; font-size: 14px;">This link will expire in 24 hours.</p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
    <p style="color: #999; font-size: 12px;">If you didn't create an account, you can safely ignore this email.</p>
</body>
</html>
`, verifyURL, verifyURL, verifyURL)

	_, err := s.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    s.fromAddress,
		To:      []string{to},
		Subject: "Verify your email - AgentsMesh",
		Html:    html,
	})
	return err
}

// SendPasswordResetEmail sends password reset email via Resend
func (s *ResendService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Reset your password</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #333;">Reset your password</h1>
    <p style="color: #666; font-size: 16px;">We received a request to reset your password. Click the button below to proceed:</p>
    <p style="margin: 30px 0;">
        <a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 500;">Reset Password</a>
    </p>
    <p style="color: #999; font-size: 14px;">Or copy this link: <a href="%s" style="color: #0070f3;">%s</a></p>
    <p style="color: #999; font-size: 14px;">This link will expire in 1 hour.</p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
    <p style="color: #999; font-size: 12px;">If you didn't request a password reset, you can safely ignore this email.</p>
</body>
</html>
`, resetURL, resetURL, resetURL)

	_, err := s.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    s.fromAddress,
		To:      []string{to},
		Subject: "Reset your password - AgentsMesh",
		Html:    html,
	})
	return err
}

// SendOrgInvitationEmail sends organization invitation via Resend
func (s *ResendService) SendOrgInvitationEmail(ctx context.Context, to, orgName, inviterName, token string) error {
	inviteURL := fmt.Sprintf("%s/invite/%s", s.baseURL, token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>You've been invited to join %s</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #333;">You're invited!</h1>
    <p style="color: #666; font-size: 16px;"><strong>%s</strong> has invited you to join <strong>%s</strong> on AgentsMesh.</p>
    <p style="margin: 30px 0;">
        <a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 500;">Accept Invitation</a>
    </p>
    <p style="color: #999; font-size: 14px;">Or copy this link: <a href="%s" style="color: #0070f3;">%s</a></p>
    <p style="color: #999; font-size: 14px;">This invitation will expire in 7 days.</p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
    <p style="color: #999; font-size: 12px;">If you don't want to join, you can safely ignore this email.</p>
</body>
</html>
`, orgName, inviterName, orgName, inviteURL, inviteURL, inviteURL)

	_, err := s.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    s.fromAddress,
		To:      []string{to},
		Subject: fmt.Sprintf("You've been invited to join %s - AgentsMesh", orgName),
		Html:    html,
	})
	return err
}

// SendRenewalReminder sends subscription renewal reminder via Resend
func (s *ResendService) SendRenewalReminder(ctx context.Context, to, orgName, planName string, expiryDate time.Time, daysRemaining int, orgSlug string) error {
	renewURL := fmt.Sprintf("%s/%s/settings?scope=organization&tab=billing", s.baseURL, orgSlug)
	expiryDateStr := expiryDate.Format("2006-01-02")

	var urgencyClass, urgencyText string
	switch {
	case daysRemaining <= 1:
		urgencyClass = "color: #dc2626;" // red
		urgencyText = "Your subscription expires tomorrow!"
	case daysRemaining <= 3:
		urgencyClass = "color: #ea580c;" // orange
		urgencyText = fmt.Sprintf("Your subscription expires in %d days", daysRemaining)
	default:
		urgencyClass = "color: #ca8a04;" // yellow
		urgencyText = fmt.Sprintf("Your subscription expires in %d days", daysRemaining)
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Subscription Renewal Reminder</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #333;">Subscription Renewal Reminder</h1>
    <p style="%s font-size: 18px; font-weight: 600;">%s</p>
    <p style="color: #666; font-size: 16px;">Your <strong>%s</strong> plan for <strong>%s</strong> will expire on <strong>%s</strong>.</p>
    <p style="color: #666; font-size: 16px;">To continue using all features without interruption, please renew your subscription before the expiry date.</p>
    <p style="margin: 30px 0;">
        <a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 500;">Renew Now</a>
    </p>
    <p style="color: #999; font-size: 14px;">Or visit your billing settings: <a href="%s" style="color: #0070f3;">%s</a></p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
    <p style="color: #999; font-size: 12px;">If you don't renew, your organization will be frozen and you won't be able to create new pods or invite members. Your data will be preserved.</p>
</body>
</html>
`, urgencyClass, urgencyText, planName, orgName, expiryDateStr, renewURL, renewURL, renewURL)

	_, err := s.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    s.fromAddress,
		To:      []string{to},
		Subject: fmt.Sprintf("Subscription Renewal Reminder - %s expires in %d days", orgName, daysRemaining),
		Html:    html,
	})
	return err
}

// ConsoleService implements email service for development (prints to console)
type ConsoleService struct {
	baseURL string
}

// SendVerificationEmail prints verification email to console
func (s *ConsoleService) SendVerificationEmail(ctx context.Context, to, token string) error {
	verifyURL := fmt.Sprintf("%s/verify-email/callback?token=%s", s.baseURL, token)
	slog.Info("console email: verification",
		"to", to, "verify_url", verifyURL)
	return nil
}

// SendPasswordResetEmail prints password reset email to console
func (s *ConsoleService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)
	slog.Info("console email: password reset",
		"to", to, "reset_url", resetURL)
	return nil
}

// SendOrgInvitationEmail prints organization invitation to console
func (s *ConsoleService) SendOrgInvitationEmail(ctx context.Context, to, orgName, inviterName, token string) error {
	inviteURL := fmt.Sprintf("%s/invite/%s", s.baseURL, token)
	slog.Info("console email: organization invitation",
		"to", to, "org", orgName, "inviter", inviterName, "invite_url", inviteURL)
	return nil
}

// SendRenewalReminder prints renewal reminder to console
func (s *ConsoleService) SendRenewalReminder(ctx context.Context, to, orgName, planName string, expiryDate time.Time, daysRemaining int, orgSlug string) error {
	renewURL := fmt.Sprintf("%s/%s/settings?scope=organization&tab=billing", s.baseURL, orgSlug)
	slog.Info("console email: renewal reminder",
		"to", to, "org", orgName, "plan", planName,
		"expiry", expiryDate.Format("2006-01-02"), "days_remaining", daysRemaining,
		"renew_url", renewURL)
	return nil
}

// SMTPService implements email service via SMTP
type SMTPService struct {
	host     string
	port     int
	username string
	password string
	from     string
	baseURL  string
}

// SendVerificationEmail sends email verification via SMTP
func (s *SMTPService) SendVerificationEmail(ctx context.Context, to, token string) error {
	verifyURL := fmt.Sprintf("%s/verify-email/callback?token=%s", s.baseURL, token)

	subject := "Verify your email - AgentsMesh"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Verify your email</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #333;">Welcome to AgentsMesh!</h1>
    <p style="color: #666; font-size: 16px;">Please verify your email address by clicking the button below:</p>
    <p style="margin: 30px 0;">
        <a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 500;">Verify Email</a>
    </p>
    <p style="color: #999; font-size: 14px;">Or copy this link: <a href="%s" style="color: #0070f3;">%s</a></p>
    <p style="color: #999; font-size: 14px;">This link will expire in 24 hours.</p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
    <p style="color: #999; font-size: 12px;">If you didn't create an account, you can safely ignore this email.</p>
</body>
</html>
`, verifyURL, verifyURL, verifyURL)

	return s.sendEmail(to, subject, body)
}

// SendPasswordResetEmail sends password reset email via SMTP
func (s *SMTPService) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

	subject := "Reset your password - AgentsMesh"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Reset your password</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #333;">Reset your password</h1>
    <p style="color: #666; font-size: 16px;">We received a request to reset your password. Click the button below to proceed:</p>
    <p style="margin: 30px 0;">
        <a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 500;">Reset Password</a>
    </p>
    <p style="color: #999; font-size: 14px;">Or copy this link: <a href="%s" style="color: #0070f3;">%s</a></p>
    <p style="color: #999; font-size: 14px;">This link will expire in 1 hour.</p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
    <p style="color: #999; font-size: 12px;">If you didn't request a password reset, you can safely ignore this email.</p>
</body>
</html>
`, resetURL, resetURL, resetURL)

	return s.sendEmail(to, subject, body)
}

// SendOrgInvitationEmail sends organization invitation via SMTP
func (s *SMTPService) SendOrgInvitationEmail(ctx context.Context, to, orgName, inviterName, token string) error {
	inviteURL := fmt.Sprintf("%s/invite/%s", s.baseURL, token)

	subject := fmt.Sprintf("You've been invited to join %s - AgentsMesh", orgName)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>You've been invited to join %s</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #333;">You're invited!</h1>
    <p style="color: #666; font-size: 16px;"><strong>%s</strong> has invited you to join <strong>%s</strong> on AgentsMesh.</p>
    <p style="margin: 30px 0;">
        <a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 500;">Accept Invitation</a>
    </p>
    <p style="color: #999; font-size: 14px;">Or copy this link: <a href="%s" style="color: #0070f3;">%s</a></p>
    <p style="color: #999; font-size: 14px;">This invitation will expire in 7 days.</p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
    <p style="color: #999; font-size: 12px;">If you don't want to join, you can safely ignore this email.</p>
</body>
</html>
`, orgName, inviterName, orgName, inviteURL, inviteURL, inviteURL)

	return s.sendEmail(to, subject, body)
}

// SendRenewalReminder sends subscription renewal reminder via SMTP
func (s *SMTPService) SendRenewalReminder(ctx context.Context, to, orgName, planName string, expiryDate time.Time, daysRemaining int, orgSlug string) error {
	renewURL := fmt.Sprintf("%s/%s/settings?scope=organization&tab=billing", s.baseURL, orgSlug)
	expiryDateStr := expiryDate.Format("2006-01-02")

	var urgencyClass, urgencyText string
	switch {
	case daysRemaining <= 1:
		urgencyClass = "color: #dc2626;" // red
		urgencyText = "Your subscription expires tomorrow!"
	case daysRemaining <= 3:
		urgencyClass = "color: #ea580c;" // orange
		urgencyText = fmt.Sprintf("Your subscription expires in %d days", daysRemaining)
	default:
		urgencyClass = "color: #ca8a04;" // yellow
		urgencyText = fmt.Sprintf("Your subscription expires in %d days", daysRemaining)
	}

	subject := fmt.Sprintf("Subscription Renewal Reminder - %s expires in %d days", orgName, daysRemaining)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Subscription Renewal Reminder</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #333;">Subscription Renewal Reminder</h1>
    <p style="%s font-size: 18px; font-weight: 600;">%s</p>
    <p style="color: #666; font-size: 16px;">Your <strong>%s</strong> plan for <strong>%s</strong> will expire on <strong>%s</strong>.</p>
    <p style="color: #666; font-size: 16px;">To continue using all features without interruption, please renew your subscription before the expiry date.</p>
    <p style="margin: 30px 0;">
        <a href="%s" style="background-color: #0070f3; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: 500;">Renew Now</a>
    </p>
    <p style="color: #999; font-size: 14px;">Or visit your billing settings: <a href="%s" style="color: #0070f3;">%s</a></p>
    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
    <p style="color: #999; font-size: 12px;">If you don't renew, your organization will be frozen and you won't be able to create new pods or invite members. Your data will be preserved.</p>
</body>
</html>
`, urgencyClass, urgencyText, planName, orgName, expiryDateStr, renewURL, renewURL, renewURL)

	return s.sendEmail(to, subject, body)
}

// sendEmail sends an email via SMTP
func (s *SMTPService) sendEmail(to, subject, htmlBody string) error {
	// Parse from address to extract email and name
	from := s.from
	if from == "" {
		from = s.username
	}

	// Build message
	msg := s.buildMessage(from, to, subject, htmlBody)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	// For port 465, try TLS connection first (implicit TLS/SSL)
	if s.port == 465 {
		return s.sendWithTLS(addr, from, to, msg)
	}

	// For other ports (25, 587), use STARTTLS
	return s.sendWithSTARTTLS(addr, from, to, msg)
}

// sendWithTLS sends email using direct TLS connection (for port 465)
func (s *SMTPService) sendWithTLS(addr, from, to, msg string) error {
	tlsConfig := &tls.Config{
		ServerName: s.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server via TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate if credentials are provided
	if s.username != "" && s.password != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	return s.sendMessage(client, from, to, msg)
}

// sendWithSTARTTLS sends email using STARTTLS (for ports 25, 587)
func (s *SMTPService) sendWithSTARTTLS(addr, from, to, msg string) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Send EHLO/HELO first
	if err := client.Hello(s.host); err != nil {
		return fmt.Errorf("SMTP HELLO failed: %w", err)
	}

	// Check if STARTTLS is supported and use it
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: s.host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
	}

	// Authenticate if credentials are provided
	// Must authenticate AFTER STARTTLS for Outlook/Office365
	if s.username != "" && s.password != "" {
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	return s.sendMessage(client, from, to, msg)
}

// sendMessage sends the actual email message using an established SMTP client
func (s *SMTPService) sendMessage(client *smtp.Client, from, to, msg string) error {
	// Send MAIL command
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL command failed: %w", err)
	}

	// Send RCPT command
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT command failed: %w", err)
	}

	// Send DATA command and message body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA command failed: %w", err)
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close email body: %w", err)
	}

	return client.Quit()
}

// buildMessage builds RFC 5322 compliant email message
func (s *SMTPService) buildMessage(from, to, subject, htmlBody string) string {
	var msg strings.Builder

	fmt.Fprintf(&msg, "From: %s\r\n", from)
	fmt.Fprintf(&msg, "To: %s\r\n", to)
	fmt.Fprintf(&msg, "Subject: %s\r\n", subject)
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	return msg.String()
}
