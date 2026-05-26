package contract

import "errors"

// Общие ошибки, которые могут быть возвращены
var (
	// ErrEmptyQuery - пустой поисковый запрос
	ErrEmptyQuery = errors.New("query is empty")

	// ErrEmptyUserID - пустой идентификатор пользователя
	ErrEmptyUserID = errors.New("user_id is empty")

	// ErrInvalidTimestamp - некорректный timestamp
	ErrInvalidTimestamp = errors.New("timestamp is invalid")
)
