package services

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SecurityLogService provides security-focused logging capabilities
type SecurityLogService struct {
	serviceName  string
	hostname     string
	processID    int
	outputChan   chan *SecurityLogEntry
	shutdown     chan struct{}
	workers      int
	redactFields []string // Fields to redact in logs (PII)
}

// SecurityLogEntry represents a security-focused log entry
type SecurityLogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Service     string                 `json:"service"`
	Event       string                 `json:"event"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	ClientIP    string                 `json:"client_ip,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Path        string                 `json:"path,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	StatusCode  int                    `json:"status_code,omitempty"`
	Hostname    string                 `json:"hostname,omitempty"`
	ProcessID   int                    `json:"process_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	ThreatScore int                    `json:"threat_score,omitempty"`
	ThreatType  string                 `json:"threat_type,omitempty"`
}

// NewSecurityLogService creates a new security logging service
func NewSecurityLogService(serviceName string) *SecurityLogService {
	hostname, _ := os.Hostname()

	securityLogger := &SecurityLogService{
		serviceName:  serviceName,
		hostname:     hostname,
		processID:    os.Getpid(),
		outputChan:   make(chan *SecurityLogEntry, 500), // Buffer 500 security log entries
		shutdown:     make(chan struct{}),
		workers:      2, // 2 concurrent workers for security logs
		redactFields: []string{"password", "token", "apiKey", "secret", "key", "credential", "authorization", "cookie", "email", "phone", "ssn", "credit_card"},
	}

	// Start worker goroutines for async logging
	for i := 0; i < securityLogger.workers; i++ {
		go securityLogger.worker(i)
	}

	return securityLogger
}

// worker processes security log entries from the channel
func (s *SecurityLogService) worker(workerID int) {
	for {
		select {
		case logEntry, ok := <-s.outputChan:
			if !ok {
				return
			}

			// Redact sensitive information
			s.redactSensitiveData(logEntry)

			// Output to stdout as JSON
			jsonData, err := json.Marshal(logEntry)
			if err != nil {
				// Fallback to basic output if JSON marshaling fails
				fmt.Fprintf(os.Stderr, "SECURITY_LOG_ERROR: Failed to marshal log entry: %v\n", err)
				continue
			}

			fmt.Fprintln(os.Stdout, string(jsonData))

		case <-s.shutdown:
			return
		}
	}
}

// redactSensitiveData removes sensitive information from log entries
func (s *SecurityLogService) redactSensitiveData(logEntry *SecurityLogEntry) {
	if logEntry.Data != nil {
		for key, value := range logEntry.Data {
			for _, redactField := range s.redactFields {
				if strings.Contains(strings.ToLower(key), strings.ToLower(redactField)) {
					logEntry.Data[key] = "[REDACTED]"
					break
				}
			}

			// If the value is a string that might contain sensitive data
			if strValue, ok := value.(string); ok {
				for _, redactField := range s.redactFields {
					if strings.Contains(strings.ToLower(strValue), strings.ToLower(redactField)) {
						logEntry.Data[key] = "[REDACTED]"
						break
					}
				}
			}
		}
	}
}

// LogSecurityEvent logs a security-related event
func (s *SecurityLogService) LogSecurityEvent(event, message string, data map[string]interface{}) {
	logEntry := &SecurityLogEntry{
		Timestamp: time.Now().UTC(),
		Level:     "SECURITY",
		Service:   s.serviceName,
		Event:     event,
		Message:   message,
		Data:      data,
		Hostname:  s.hostname,
		ProcessID: s.processID,
	}

	select {
	case s.outputChan <- logEntry:
		// Successfully queued
	default:
		// Channel full - log to stderr as fallback
		jsonData, _ := json.Marshal(logEntry)
		fmt.Fprintf(os.Stderr, "SECURITY_LOGGER_OVERFLOW: %s\n", string(jsonData))
	}
}

// LogSecurityEventWithContext logs a security event with request context
func (s *SecurityLogService) LogSecurityEventWithContext(c *fiber.Ctx, event, message string, data map[string]interface{}) {
	logEntry := &SecurityLogEntry{
		Timestamp:  time.Now().UTC(),
		Level:      "SECURITY",
		Service:    s.serviceName,
		Event:      event,
		Message:    message,
		Data:       data,
		ClientIP:   c.IP(),
		Method:     c.Method(),
		Path:       c.Path(),
		UserAgent:  c.Get("User-Agent"),
		StatusCode: c.Response().StatusCode(),
		Hostname:   s.hostname,
		ProcessID:  s.processID,
	}

	// Add user ID if available in context
	if userID := c.Locals("userID"); userID != nil {
		if idStr, ok := userID.(string); ok {
			logEntry.UserID = idStr
		}
	}

	// Add session ID if available
	if sessionID := c.Locals("sessionID"); sessionID != nil {
		if idStr, ok := sessionID.(string); ok {
			logEntry.SessionID = idStr
		}
	}

	select {
	case s.outputChan <- logEntry:
		// Successfully queued
	default:
		// Channel full - log to stderr as fallback
		jsonData, _ := json.Marshal(logEntry)
		fmt.Fprintf(os.Stderr, "SECURITY_LOGGER_OVERFLOW: %s\n", string(jsonData))
	}
}

// LogSuspiciousActivity logs potentially suspicious activity
func (s *SecurityLogService) LogSuspiciousActivity(activityType, description string, ip string, userID string, additionalData map[string]interface{}) {
	data := map[string]interface{}{
		"activity_type": activityType,
		"description":   description,
		"ip_address":    ip,
		"user_id":       userID,
	}

	for k, v := range additionalData {
		data[k] = v
	}

	s.LogSecurityEvent("suspicious_activity", description, data)
}

// LogBruteForceAttempt logs potential brute force attempts
func (s *SecurityLogService) LogBruteForceAttempt(ip string, path string, userAgent string) {
	data := map[string]interface{}{
		"ip_address":   ip,
		"path":         path,
		"user_agent":   userAgent,
		"attempt_type": "brute_force",
	}

	s.LogSecurityEvent("brute_force_attempt", "Potential brute force attack detected", data)
}

// LogSQLInjectionAttempt logs potential SQL injection attempts
func (s *SecurityLogService) LogSQLInjectionAttempt(ip string, path string, query string, matchedPattern string) {
	data := map[string]interface{}{
		"ip_address":       ip,
		"path":             path,
		"suspicious_query": query,
		"matched_pattern":  matchedPattern,
		"attack_type":      "sql_injection",
	}

	s.LogSecurityEvent("sql_injection_attempt", "Potential SQL injection detected", data)
}

// LogXSSAttempt logs potential XSS attempts
func (s *SecurityLogService) LogXSSAttempt(ip string, path string, input string) {
	data := map[string]interface{}{
		"ip_address":       ip,
		"path":             path,
		"suspicious_input": input,
		"attack_type":      "xss",
	}

	s.LogSecurityEvent("xss_attempt", "Potential XSS attack detected", data)
}

// LogUnauthorizedAccess logs unauthorized access attempts
func (s *SecurityLogService) LogUnauthorizedAccess(ip string, path string, userID string, resource string) {
	data := map[string]interface{}{
		"ip_address":  ip,
		"path":        path,
		"user_id":     userID,
		"resource":    resource,
		"access_type": "unauthorized",
	}

	s.LogSecurityEvent("unauthorized_access", "Unauthorized access attempt", data)
}

// LogDataExfiltrationAttempt logs potential data exfiltration attempts
func (s *SecurityLogService) LogDataExfiltrationAttempt(ip string, userID string, query string, dataSize int64) {
	data := map[string]interface{}{
		"ip_address":      ip,
		"user_id":         userID,
		"query":           query,
		"data_size_bytes": dataSize,
		"attempt_type":    "data_exfiltration",
	}

	s.LogSecurityEvent("data_exfiltration_attempt", "Potential data exfiltration detected", data)
}

// Stop gracefully shuts down the security logger service
func (s *SecurityLogService) Stop() {
	close(s.shutdown)
	close(s.outputChan)
}

// Global security logger instance
var GlobalSecurityLogger *SecurityLogService

// InitSecurityLogger initializes the global security logger
func InitSecurityLogger(serviceName string) {
	GlobalSecurityLogger = NewSecurityLogService(serviceName)
}

// Helper functions for global security logger access
func LogSecurityEvent(event, message string, data map[string]interface{}) {
	if GlobalSecurityLogger != nil {
		GlobalSecurityLogger.LogSecurityEvent(event, message, data)
	}
}

func LogSecurityEventWithContext(c *fiber.Ctx, event, message string, data map[string]interface{}) {
	if GlobalSecurityLogger != nil {
		GlobalSecurityLogger.LogSecurityEventWithContext(c, event, message, data)
	}
}

func LogSuspiciousActivity(activityType, description string, ip string, userID string, additionalData map[string]interface{}) {
	if GlobalSecurityLogger != nil {
		GlobalSecurityLogger.LogSuspiciousActivity(activityType, description, ip, userID, additionalData)
	}
}

func LogBruteForceAttempt(ip string, path string, userAgent string) {
	if GlobalSecurityLogger != nil {
		GlobalSecurityLogger.LogBruteForceAttempt(ip, path, userAgent)
	}
}

func LogSQLInjectionAttempt(ip string, path string, query string, matchedPattern string) {
	if GlobalSecurityLogger != nil {
		GlobalSecurityLogger.LogSQLInjectionAttempt(ip, path, query, matchedPattern)
	}
}

func LogXSSAttempt(ip string, path string, input string) {
	if GlobalSecurityLogger != nil {
		GlobalSecurityLogger.LogXSSAttempt(ip, path, input)
	}
}

func LogUnauthorizedAccess(ip string, path string, userID string, resource string) {
	if GlobalSecurityLogger != nil {
		GlobalSecurityLogger.LogUnauthorizedAccess(ip, path, userID, resource)
	}
}

func LogDataExfiltrationAttempt(ip string, userID string, query string, dataSize int64) {
	if GlobalSecurityLogger != nil {
		GlobalSecurityLogger.LogDataExfiltrationAttempt(ip, userID, query, dataSize)
	}
}
