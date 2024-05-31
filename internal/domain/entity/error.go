package entity

type AppError struct {
	HTTPStatusCode int
	Message        string
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrUserNotFound *AppError = &AppError{
		HTTPStatusCode: 404,
		Message:        "user not found",
	}
	ErrValidation *AppError = &AppError{
		HTTPStatusCode: 400,
		Message:        "validation error",
	}
	ErrTokenExpired *AppError = &AppError{
		HTTPStatusCode: 401,
		Message:        "token expired",
	}
)
