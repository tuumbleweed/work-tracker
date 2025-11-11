// you can add any code you want here but don't commit it.
// keep it empty for future projects and for use as a template.
package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
	"mime"

	tl "github.com/tuumbleweed/tintlog/logger"
	"github.com/tuumbleweed/tintlog/palette"
	"github.com/tuumbleweed/xerr"

	"work-tracker/src/pkg/config"
	"work-tracker/src/pkg/util"
)

// --- multi-flag support for -to ---
type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	if v == "" {
		return nil
	}
	// allow comma-separated or repeated -to
	parts := strings.Split(v, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			*m = append(*m, p)
		}
	}
	return nil
}

func main() {
	util.CheckIfEnvVarsPresent([]string{})

	// ---------- Common flags ----------
	configPath := flag.String("config", "./cfg/config.json", "Path to configuration file.")

	// ---------- Program flags ----------
	from := flag.String("from", "", "From email address (required).")
	var to multiFlag
	flag.Var(&to, "to", "Recipient email (repeatable or comma-separated) (required).")
	subject := flag.String("subject", "Test email from Go", "Email subject.")
	htmlPath := flag.String("html", "./tmp/email.html", "Path to HTML body file.")
	textPath := flag.String("text", "./tmp/email.txt", "Path to plain text body file.")

	smtpHost := flag.String("smtp-host", "smtp-relay.gmail.com", "SMTP relay host (e.g., smtp-relay.gmail.com or smtp.gmail.com).")
	smtpPort := flag.Int("smtp-port", 587, "SMTP port (587 for STARTTLS, 465 for SMTPS).")

	smtpUser := flag.String("smtp-user", "", "SMTP username (omit for IP-allowed relays).")
	smtpPass := flag.String("smtp-pass", "", "SMTP password or app password.")
	noAuth := flag.Bool("no-auth", false, "Disable SMTP authentication (for IP-allowed relays).")

	useSMTPS := flag.Bool("smtps", false, "Use SMTPS (implicit TLS, typical on port 465).")
	startTLS := flag.Bool("starttls", true, "Use STARTTLS upgrade on plain connection (port 587). Ignored if -smtps is true.")
	skipVerify := flag.Bool("skip-verify", false, "Skip TLS certificate verification (INSECURE).")

	ehloName := flag.String("ehlo", "", "EHLO/HELO name. Defaults to system hostname.")

	// Parse and init config
	flag.Parse()
	config.InitializeConfig(*configPath)

	// ---- validations ----
	if *from == "" {
		xerr.QuitIfError(fmt.Errorf("missing -from"), "Missing -from address")
	}
	if len(to) == 0 {
		xerr.QuitIfError(fmt.Errorf("missing -to"), "At least one -to address is required")
	}
	if *smtpHost == "" || *smtpPort <= 0 {
		xerr.QuitIfError(fmt.Errorf("smtp invalid"), "Invalid SMTP host/port")
	}
	if !*noAuth && (*smtpUser == "" || *smtpPass == "") {
		tl.Log(tl.Warning, palette.Yellow, "Auth enabled but -smtp-user or -smtp-pass is empty. Use -no-auth for IP-allowed relays.")
	}

	if *useSMTPS && *startTLS {
		tl.Log(tl.Info, palette.Blue, "Note: -smtps=true overrides -starttls (implicit TLS will be used).")
	}

	// EHLO name
	if *ehloName == "" {
		if hn, err := os.Hostname(); err == nil && hn != "" {
			*ehloName = hn
		} else {
			*ehloName = "localhost"
		}
	}

	// ---- load bodies from files (at least one required) ----
	textBody, _ := readFileIfExists(*textPath)
	htmlBody, _ := readFileIfExists(*htmlPath)
	if strings.TrimSpace(textBody) == "" && strings.TrimSpace(htmlBody) == "" {
		xerr.QuitIfError(fmt.Errorf("no bodies"), "Neither -text nor -html file has content")
	}

	// ---- build MIME message ----
	msg, err, emsg := buildMessage(*from, to, *subject, textBody, htmlBody)
	xerr.QuitIfError(err, emsg)

	// ---- send via SMTP relay ----
	err, emsg = sendSMTP(*smtpHost, *smtpPort, *ehloName, *useSMTPS, *startTLS, *skipVerify, !*noAuth, *smtpUser, *smtpPass, *from, to, msg)
	xerr.QuitIfError(err, emsg)

	tl.Log(
		tl.Notice, palette.GreenBold,
		"Email sent successfully via %s:%d to %s",
		*smtpHost, *smtpPort, strings.Join(to, ", "),
	)
}

// readFileIfExists returns file contents as UTF-8 string (or empty string if file doesn't exist).
func readFileIfExists(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		// If not found, just return empty (caller decides)
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}

func buildMessage(from string, to []string, subject, textBody, htmlBody string) ([]byte, error, string) {
	var buf bytes.Buffer

	// Headers
	now := time.Now().UTC()
	msgID := makeMessageID()

	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(to, ", "))
	fmt.Fprintf(&buf, "Subject: %s\r\n", encodeHeader(subject))
	fmt.Fprintf(&buf, "Date: %s\r\n", now.Format(time.RFC1123Z))
	fmt.Fprintf(&buf, "Message-ID: <%s>\r\n", msgID)
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")

	// Select appropriate body structure
	textBody = strings.ReplaceAll(textBody, "\r\n", "\n")
	htmlBody = strings.ReplaceAll(htmlBody, "\r\n", "\n")

	switch {
	case strings.TrimSpace(textBody) != "" && strings.TrimSpace(htmlBody) != "":
		// multipart/alternative: text then html
		boundary := randomBoundary("alt")
		fmt.Fprintf(&buf, "Content-Type: multipart/alternative; boundary=%q\r\n", boundary)
		fmt.Fprintf(&buf, "\r\n")
		// text/plain
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
		fmt.Fprintf(&buf, "Content-Transfer-Encoding: 8bit\r\n\r\n")
		writeWithCRLF(&buf, textBody)
		fmt.Fprintf(&buf, "\r\n")
		// text/html
		fmt.Fprintf(&buf, "--%s\r\n", boundary)
		fmt.Fprintf(&buf, "Content-Type: text/html; charset=UTF-8\r\n")
		fmt.Fprintf(&buf, "Content-Transfer-Encoding: 8bit\r\n\r\n")
		writeWithCRLF(&buf, htmlBody)
		fmt.Fprintf(&buf, "\r\n--%s--\r\n", boundary)

	case strings.TrimSpace(htmlBody) != "":
		fmt.Fprintf(&buf, "Content-Type: text/html; charset=UTF-8\r\n")
		fmt.Fprintf(&buf, "Content-Transfer-Encoding: 8bit\r\n\r\n")
		writeWithCRLF(&buf, htmlBody)

	default:
		// text only
		fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
		fmt.Fprintf(&buf, "Content-Transfer-Encoding: 8bit\r\n\r\n")
		writeWithCRLF(&buf, textBody)
	}

	return buf.Bytes(), nil, "Failed to build MIME message"
}

func sendSMTP(
	host string, port int, ehloName string,
	useSMTPS, useStartTLS, skipVerify, doAuth bool,
	user, pass, from string, recipients []string,
	msg []byte,
) (error, string) {

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	tlsCfg := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: skipVerify, // NOT recommended; only for debugging
		MinVersion:         tls.VersionTLS12,
	}

	var (
		c   *smtp.Client
		err error
	)

	if useSMTPS {
		// Implicit TLS (465)
		conn, dialErr := tls.Dial("tcp", addr, tlsCfg)
		if dialErr != nil {
			return dialErr, "TLS dial (SMTPS) failed"
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			return err, "SMTP client creation (SMTPS) failed"
		}
	} else {
		// Plain TCP (587), then STARTTLS
		conn, dErr := net.Dial("tcp", addr)
		if dErr != nil {
			return dErr, "TCP dial failed"
		}
		c, err = smtp.NewClient(conn, host)
		if err != nil {
			return err, "SMTP client creation failed"
		}
	}

	// --- EHLO exactly once (required before Extension/StartTLS/Auth) ---
	if err = c.Hello(ehloName); err != nil {
		_ = c.Close()
		return err, "SMTP EHLO/HELO failed"
	}

	// --- STARTTLS if requested (and only on non-SMTPS path) ---
	if !useSMTPS && useStartTLS {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			_ = c.Close()
			return fmt.Errorf("server does not advertise STARTTLS"), "STARTTLS not supported"
		}
		if err = c.StartTLS(tlsCfg); err != nil {
			_ = c.Close()
			return err, "STARTTLS upgrade failed"
		}
		// IMPORTANT: Do NOT call Hello() again here; net/smtp forbids it.
	}

	// --- AUTH (optional) ---
	if doAuth {
		auth := smtp.PlainAuth("", user, pass, host)
		if err = c.Auth(auth); err != nil {
			_ = c.Close()
			return err, "SMTP AUTH failed"
		}
	}

	// --- MAIL FROM / RCPT TO / DATA ---
	if err = c.Mail(from); err != nil {
		_ = c.Close()
		return err, "SMTP MAIL FROM failed"
	}
	for _, r := range recipients {
		if err = c.Rcpt(r); err != nil {
			_ = c.Close()
			return err, fmt.Sprintf("SMTP RCPT TO failed for %s", r)
		}
	}
	w, err := c.Data()
	if err != nil {
		_ = c.Close()
		return err, "SMTP DATA command failed"
	}
	if _, err = w.Write(msg); err != nil {
		_ = w.Close()
		_ = c.Close()
		return err, "Writing message body failed"
	}
	if err = w.Close(); err != nil {
		_ = c.Close()
		return err, "Closing DATA writer failed"
	}
	if err = c.Quit(); err != nil {
		_ = c.Close()
		return err, "SMTP QUIT failed"
	}
	return nil, ""
}

func randomBoundary(prefix string) string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().UnixNano(), hex.EncodeToString(b[:]))
}

func encodeHeader(s string) string {
	// Encode non-ASCII subjects safely
	if isASCII(s) {
		return s
	}
	return mime.QEncoding.Encode("utf-8", s)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

// writeWithCRLF writes text with \n normalized to \r\n per RFC 5322
func writeWithCRLF(w io.Writer, s string) {
	// normalize lone \n to \r\n
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", "\r\n")
	_, _ = io.WriteString(w, s)
}

func makeMessageID() string {
	var r [8]byte
	_, _ = rand.Read(r[:])
	host, _ := os.Hostname()
	if host == "" {
		host = "localhost"
	}
	return fmt.Sprintf("%d.%s@%s", time.Now().UnixNano(), hex.EncodeToString(r[:]), host)
}
