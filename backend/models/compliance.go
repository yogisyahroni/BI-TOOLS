package models

import "time"

// ComplianceReport represents a comprehensive compliance report
type ComplianceReport struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	GeneratedAt time.Time `json:"generated_at"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Checks      map[string]interface{} `json:"checks" gorm:"serializer:json"`
	Summary     ComplianceSummary `json:"summary"`
	Details     map[string]interface{} `json:"details" gorm:"serializer:json"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ComplianceSummary contains high-level compliance metrics
type ComplianceSummary struct {
	TotalChecks      int `json:"total_checks"`
	PassedChecks     int `json:"passed_checks"`
	FailedChecks     int `json:"failed_checks"`
	ComplianceScore  float64 `json:"compliance_score"` // Percentage of checks passed
	OverallStatus    string `json:"overall_status"`    // COMPLIANT, NON_COMPLIANT, PARTIAL
	LastUpdated      time.Time `json:"last_updated"`
	NextAuditDue     time.Time `json:"next_audit_due"`
}

// GDPRComplianceMetrics contains GDPR-specific compliance metrics
type GDPRComplianceMetrics struct {
	RightToErasureRequests int `json:"right_to_erasure_requests"`
	SuccessfulErasures   int `json:"successful_erasures"`
	DataBreachIncidents  int `json:"data_breach_incidents"`
	DSAResponses         int `json:"dsa_responses"` // Data Subject Access Requests
	ConsentManagement    ConsentMetrics `json:"consent_management"`
}

// ConsentMetrics tracks consent-related metrics
type ConsentMetrics struct {
	TotalConsents       int `json:"total_consents"`
	ActiveConsents      int `json:"active_consents"`
	WithdrawnConsents   int `json:"withdrawn_consents"`
	ExpiredConsents     int `json:"expired_consents"`
	ConsentRate         float64 `json:"consent_rate"`
}

// HIPAAComplianceMetrics contains HIPAA-specific compliance metrics
type HIPAAComplianceMetrics struct {
	AccessLogsReviewed   int `json:"access_logs_reviewed"`
	UnauthorizedAccess   int `json:"unauthorized_access"`
	PrivacyIncidents     int `json:"privacy_incidents"`
	TrainingCompletions  int `json:"training_completions"`
	AccessControlsAudited int `json:"access_controls_audited"`
}

// SOXComplianceMetrics contains SOX-specific compliance metrics
type SOXComplianceMetrics struct {
	FinancialControlsTested int `json:"financial_controls_tested"`
	ControlDeficiencies   int `json:"control_deficiencies"`
	AccessReviews         int `json:"access_reviews"`
	ChangeApprovals       int `json:"change_approvals"`
	FinancialReporting   bool `json:"financial_reporting"`
}