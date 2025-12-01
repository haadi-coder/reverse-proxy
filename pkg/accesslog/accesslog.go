package accesslog

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

type Format string

const (
	CommonFormat   Format = "common"
	JSONFormat     Format = "json"
	CombinedFormat Format = "combined"
)

type AccessLogConfig struct {
	Format Format
	Output io.Writer
}

type AccessLogger struct {
	cfg *AccessLogConfig
}

func NewLogger(cfg *AccessLogConfig) *AccessLogger {
	output := cfg.Output

	if output == nil {
		output = os.Stdout
	}

	return &AccessLogger{
		cfg: &AccessLogConfig{
			Format: cfg.Format,
			Output: output,
		},
	}
}

func (l *AccessLogger) Log(req *http.Request, resp *ResponseWriter, startTime time.Time) error {
	var line string
	entry := l.buildEntry(req, resp, startTime)

	var err error
	switch l.cfg.Format {
	case JSONFormat:
		line, err = l.formatJSON(entry)
		if err != nil {
			return fmt.Errorf("failed to format accessLog to JSON: %w", err)
		}

	case CombinedFormat:
		line = l.formatCombined(entry)
	default:
		line = l.formatCommon(entry)
	}

	_, err = l.cfg.Output.Write([]byte(line))
	if err != nil {
		return err
	}

	return nil
}

type Entry struct {
	Time       time.Time `json:"time"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Protocol   string    `json:"protocol"`
	Status     int       `json:"status"`
	Size       int       `json:"size"`
	IP         string    `json:"ip"`
	Username   string    `json:"username,omitempty"`
	Referer    string    `json:"referer,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	DurationMs int64     `json:"duration_ms,omitempty"`
}

func (l *AccessLogger) buildEntry(req *http.Request, resp *ResponseWriter, startTime time.Time) *Entry {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}

	entry := &Entry{
		Time:       startTime,
		Method:     req.Method,
		Path:       req.RequestURI,
		Protocol:   req.Proto,
		Status:     resp.StatusCode,
		Size:       resp.ContentLength,
		IP:         host,
		DurationMs: time.Since(startTime).Milliseconds(),
	}

	if username, _, ok := req.BasicAuth(); ok {
		entry.Username = username
	}

	if l.cfg.Format == CombinedFormat || l.cfg.Format == JSONFormat {
		entry.Referer = req.Referer()
		entry.UserAgent = req.UserAgent()
	}

	return entry
}

func (l *AccessLogger) formatCommon(entry *Entry) string {
	username := entry.Username
	if username == "" {
		username = "-"
	}

	return fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %d %d\n",
		entry.IP,
		username,
		entry.Time.Format("02/Jan/2006:15:04:05 -0700"),
		entry.Method,
		entry.Path,
		entry.Protocol,
		entry.Status,
		entry.Size,
	)
}

func (l *AccessLogger) formatCombined(entry *Entry) string {
	username := entry.Username
	if username == "" {
		username = "-"
	}

	referer := entry.Referer
	if referer == "" {
		referer = "-"
	}

	userAgent := entry.UserAgent
	if userAgent == "" {
		userAgent = "-"
	}

	return fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %d %d \"%s\" \"%s\"\n",
		entry.IP,
		username,
		entry.Time.Format("02/Jan/2006:15:04:05 -0700"),
		entry.Method,
		entry.Path,
		entry.Protocol,
		entry.Status,
		entry.Size,
		referer,
		userAgent,
	)
}

func (l *AccessLogger) formatJSON(entry *Entry) (string, error) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return "", fmt.Errorf("failed to marshal access log entry: %w", err)
	}

	jsonData = append(jsonData, '\n')

	return string(jsonData), nil
}
