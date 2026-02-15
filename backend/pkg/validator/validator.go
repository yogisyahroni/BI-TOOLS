package validator

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the playground validator
type CustomValidator struct {
	validate *validator.Validate
}

var (
	instance *CustomValidator
	once     sync.Once
)

// GetValidator returns the singleton instance of CustomValidator
func GetValidator() *CustomValidator {
	once.Do(func() {
		instance = &CustomValidator{
			validate: validator.New(),
		}
	})
	return instance
}

// ValidateStruct validates a struct and returns formatted errors
func (cv *CustomValidator) ValidateStruct(s interface{}) error {
	err := cv.validate.Struct(s)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		// Detailed error formatting could be added here
		// For now, return the validation error directly
		return err
	}
	return nil
}

// ValidationErrors is a helper provided to check error type
func IsValidationErrors(err error) bool {
	_, ok := err.(validator.ValidationErrors)
	return ok
}
