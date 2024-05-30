package entity

type AppError struct {
	HTTPStatusCode int
	Message        string
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrValidation *AppError = &AppError{
		HTTPStatusCode: 400,
		Message:        "Validation error",
	}
)
