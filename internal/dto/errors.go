package dto

import "errors"

var (
	ErrEmptyQuery       = errors.New("query is empty")
	ErrEmptyUserID      = errors.New("user_id is empty")
	ErrInvalidTimestamp = errors.New("timestamp is invalid")
)
