package entity

type AppError struct {
	HTTPStatusCode int
	Message        string
}

func (e AppError) Error() string {
	return e.Message
}

var (
	ErrUserNotFound = AppError{
		HTTPStatusCode: 404,
		Message:        "user not found",
	}

	ErrValidation = AppError{
		HTTPStatusCode: 400,
		Message:        "validation error",
	}

	ErrTokenExpired = AppError{
		HTTPStatusCode: 401,
		Message:        "token expired",
	}
)
