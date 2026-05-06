package auth

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

// SendVerificationCode sends a 6-digit verification code to the given email.
func SendVerificationCode(email, code string) error {
	host := getEnv("SMTP_HOST", "smtp.qq.com")
	port := getEnv("SMTP_PORT", "587")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASSWORD")
	from := getEnv("SMTP_FROM", user)

	if user == "" || pass == "" {
		return fmt.Errorf("SMTP 未配置 (SMTP_USER/SMTP_PASSWORD 为空)")
	}

	subject := "PPT Agent 验证码"
	body := fmt.Sprintf(`您的验证码是：%s

该验证码 5 分钟内有效，请勿转发给他人。

PPT Agent`, code)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n"+
		"MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, email, subject, body)

	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", user, pass, host)

	if strings.HasPrefix(port, "465") {
		// SMTPS
		conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
		if err != nil {
			return fmt.Errorf("TLS dial: %w", err)
		}
		client, err := smtp.NewClient(conn, host)
		if err != nil {
			return fmt.Errorf("SMTP client: %w", err)
		}
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP auth: %w", err)
		}
		if err := client.Mail(from); err != nil {
			return fmt.Errorf("MAIL: %w", err)
		}
		if err := client.Rcpt(email); err != nil {
			return fmt.Errorf("RCPT: %w", err)
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("DATA: %w", err)
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
		w.Close()
		client.Quit()
		return nil
	}

	return smtp.SendMail(addr, auth, from, []string{email}, []byte(msg))
}

func getEnv(key, defaultVal string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	return v
}
