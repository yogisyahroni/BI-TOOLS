package services

import (
	"insight-engine-backend/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: AlertService — EvaluateAlert (Pure Logic, No DB)
//
// EvaluateAlert takes an alert (with Column, Operator, Threshold) and a
// result map, and returns whether the alert is triggered.
// ─────────────────────────────────────────────────────────────────────────────

func newTestAlertService() *AlertService {
	return &AlertService{}
}

// ──── Operator Tests ─────────────────────────────────────────────────────────

func TestEvaluateAlert_GreaterThan_Triggered(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "value",
		Operator:  ">",
		Threshold: 100.0,
	}
	result := map[string]interface{}{"value": 150.0}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
	assert.Equal(t, 150.0, evalResult.Value)
}

func TestEvaluateAlert_GreaterThan_NotTriggered(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "value",
		Operator:  ">",
		Threshold: 100.0,
	}
	result := map[string]interface{}{"value": 50.0}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.False(t, evalResult.Triggered)
}

func TestEvaluateAlert_LessThan_Triggered(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "cpu_usage",
		Operator:  "<",
		Threshold: 10.0,
	}
	result := map[string]interface{}{"cpu_usage": 5.0}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
}

func TestEvaluateAlert_Equal_Triggered(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "count",
		Operator:  "=",
		Threshold: 0.0,
	}
	result := map[string]interface{}{"count": 0.0}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
}

func TestEvaluateAlert_DoubleEqual_Triggered(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "count",
		Operator:  "==",
		Threshold: 42.0,
	}
	result := map[string]interface{}{"count": 42.0}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
}

func TestEvaluateAlert_GreaterEqual_Triggered(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "errors",
		Operator:  ">=",
		Threshold: 5.0,
	}
	result := map[string]interface{}{"errors": 5.0}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
}

func TestEvaluateAlert_LessEqual_Triggered(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "uptime",
		Operator:  "<=",
		Threshold: 99.0,
	}
	result := map[string]interface{}{"uptime": 98.5}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
}

func TestEvaluateAlert_NotEqual_Triggered(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "status_code",
		Operator:  "!=",
		Threshold: 200.0,
	}
	result := map[string]interface{}{"status_code": 500.0}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
}

func TestEvaluateAlert_NotEqual_NotTriggered(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "status_code",
		Operator:  "!=",
		Threshold: 200.0,
	}
	result := map[string]interface{}{"status_code": 200.0}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.False(t, evalResult.Triggered)
}

// ──── Type Conversion Tests ──────────────────────────────────────────────────

func TestEvaluateAlert_IntValue(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "count",
		Operator:  ">",
		Threshold: 10.0,
	}
	result := map[string]interface{}{"count": 42}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
	assert.Equal(t, 42.0, evalResult.Value)
}

func TestEvaluateAlert_Int32Value(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "count",
		Operator:  ">",
		Threshold: 10.0,
	}
	result := map[string]interface{}{"count": int32(42)}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
}

func TestEvaluateAlert_Int64Value(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "count",
		Operator:  ">",
		Threshold: 10.0,
	}
	result := map[string]interface{}{"count": int64(42)}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
}

func TestEvaluateAlert_Float32Value(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "percentage",
		Operator:  ">",
		Threshold: 90.0,
	}
	result := map[string]interface{}{"percentage": float32(95.5)}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
}

func TestEvaluateAlert_StringNumericValue(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "latency",
		Operator:  ">",
		Threshold: 100.0,
	}
	result := map[string]interface{}{"latency": "250.5"}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.True(t, evalResult.Triggered)
	assert.Equal(t, 250.5, evalResult.Value)
}

// ──── Error Cases ────────────────────────────────────────────────────────────

func TestEvaluateAlert_ColumnNotFound(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "nonexistent",
		Operator:  ">",
		Threshold: 100.0,
	}
	result := map[string]interface{}{"value": 150.0}

	_, err := svc.EvaluateAlert(alert, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEvaluateAlert_NonNumericString(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "name",
		Operator:  ">",
		Threshold: 100.0,
	}
	result := map[string]interface{}{"name": "not-a-number"}

	_, err := svc.EvaluateAlert(alert, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not numeric")
}

func TestEvaluateAlert_NonNumericType(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "data",
		Operator:  ">",
		Threshold: 100.0,
	}
	result := map[string]interface{}{"data": []string{"a", "b"}}

	_, err := svc.EvaluateAlert(alert, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not numeric")
}

func TestEvaluateAlert_UnknownOperator(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "value",
		Operator:  "LIKE",
		Threshold: 100.0,
	}
	result := map[string]interface{}{"value": 50.0}

	_, err := svc.EvaluateAlert(alert, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown operator")
}

// ──── Message Format Tests ───────────────────────────────────────────────────

func TestEvaluateAlert_MessageContainsColumn(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Column:    "revenue",
		Operator:  ">",
		Threshold: 1000.0,
	}
	result := map[string]interface{}{"revenue": 1500.0}

	evalResult, err := svc.EvaluateAlert(alert, result)
	assert.NoError(t, err)
	assert.Contains(t, evalResult.Message, "revenue")
}

// ─────────────────────────────────────────────────────────────────────────────
// TESTS: AlertService — CalculateNextRun (Pure Logic, No DB)
// ─────────────────────────────────────────────────────────────────────────────

func TestCalculateNextRun_CronExpression(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Schedule: "*/5 * * * *", // Every 5 minutes
		Timezone: "UTC",
	}

	nextRun, err := svc.CalculateNextRun(alert)
	assert.NoError(t, err)
	assert.NotNil(t, nextRun)
	assert.True(t, nextRun.After(time.Now()))
}

func TestCalculateNextRun_PredefinedSchedule_1m(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Schedule: "1m",
		Timezone: "UTC",
	}

	nextRun, err := svc.CalculateNextRun(alert)
	assert.NoError(t, err)
	assert.NotNil(t, nextRun)

	// Should be approximately 1 minute from now
	expectedMin := time.Now().Add(50 * time.Second)
	expectedMax := time.Now().Add(70 * time.Second)
	assert.True(t, nextRun.After(expectedMin))
	assert.True(t, nextRun.Before(expectedMax))
}

func TestCalculateNextRun_PredefinedSchedule_5m(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Schedule: "5m",
		Timezone: "UTC",
	}

	nextRun, err := svc.CalculateNextRun(alert)
	assert.NoError(t, err)
	assert.NotNil(t, nextRun)

	expectedMin := time.Now().Add(4 * time.Minute)
	assert.True(t, nextRun.After(expectedMin))
}

func TestCalculateNextRun_PredefinedSchedule_1h(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Schedule: "1h",
		Timezone: "UTC",
	}

	nextRun, err := svc.CalculateNextRun(alert)
	assert.NoError(t, err)
	assert.NotNil(t, nextRun)

	expectedMin := time.Now().Add(55 * time.Minute)
	assert.True(t, nextRun.After(expectedMin))
}

func TestCalculateNextRun_NumericMinutes(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Schedule: "10", // 10 minutes
		Timezone: "UTC",
	}

	nextRun, err := svc.CalculateNextRun(alert)
	assert.NoError(t, err)
	assert.NotNil(t, nextRun)

	expectedMin := time.Now().Add(9 * time.Minute)
	assert.True(t, nextRun.After(expectedMin))
}

func TestCalculateNextRun_InvalidSchedule(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Schedule: "invalid-schedule",
		Timezone: "UTC",
	}

	_, err := svc.CalculateNextRun(alert)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported schedule format")
}

func TestCalculateNextRun_InvalidCron(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Schedule: "* * * * * * *", // Too many fields
		Timezone: "UTC",
	}

	_, err := svc.CalculateNextRun(alert)
	assert.Error(t, err)
}

func TestCalculateNextRun_InvalidTimezone(t *testing.T) {
	svc := newTestAlertService()
	alert := &models.Alert{
		Schedule: "5m",
		Timezone: "Invalid/Timezone",
	}

	// Should fall back to UTC, not error
	nextRun, err := svc.CalculateNextRun(alert)
	assert.NoError(t, err)
	assert.NotNil(t, nextRun)
}
