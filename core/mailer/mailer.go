// Package mailer 提供 SMTP 邮件发送能力（找回密码等邮件通知）。
// 使用标准库 net/smtp，支持 SSL/TLS 与 STARTTLS。
package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"kingfisher/core/config"
)

// Mailer 封装 SMTP 配置与发送。
type Mailer struct {
	cfg config.SMTPConfig
}

// New 创建 Mailer。cfg.Enabled 为 false 时 Send 仅返回 nil（静默跳过，便于开发）。
func New(cfg config.SMTPConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

// Enabled 返回是否启用真实发送。
func (m *Mailer) Enabled() bool { return m.cfg.Enabled && m.cfg.Host != "" }

// Send 发送一封邮件。to 为收件人地址，subject 标题，body 纯文本内容。
func (m *Mailer) Send(to, subject, body string) error {
	if !m.Enabled() {
		// 未配置/禁用时静默成功（邮件仅记录，开发期不实际发送）
		return nil
	}
	from := m.cfg.From
	if from == "" {
		from = m.cfg.Username
	}
	msg := buildMessage(m.cfg.FromName, from, to, subject, body)
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)

	auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)

	if m.cfg.Port == 465 {
		// SSL/TLS 直连（QQ 465 端口）
		return sendSSL(addr, m.cfg, auth, to, msg)
	}
	// STARTTLS（QQ 587 端口）
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.StartTLS(&tls.Config{ServerName: m.cfg.Host, InsecureSkipVerify: false}); err != nil {
		return fmt.Errorf("smtp starttls: %w", err)
	}
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return client.Quit()
}

// sendSSL 通过 465 端口 SSL 直连发送。
func sendSSL(addr string, cfg config.SMTPConfig, auth smtp.Auth, to, msg string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: cfg.Host, InsecureSkipVerify: false})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	client, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(cfg.From); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return client.Quit()
}

// buildMessage 构造 RFC 5322 邮件原文。
func buildMessage(fromName, from, to, subject, body string) string {
	var b strings.Builder
	if fromName != "" {
		fmt.Fprintf(&b, "From: %s <%s>\r\n", fromName, from)
	} else {
		fmt.Fprintf(&b, "From: %s\r\n", from)
	}
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return b.String()
}
