package contract

// SearchEvent - контракт для поискового события в Kafka
// Этот контракт должны соблюдать все продюсеры, отправляющие события
type SearchEvent struct {
	// Query - поисковый запрос пользователя
	Query string `json:"query" example:"iphone 15" validate:"required"`

	// UserID - уникальный идентификатор пользователя
	UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000" validate:"required"`

	// Timestamp - время события в миллисекундах (Unix timestamp)
	Timestamp int64 `json:"timestamp" example:"1699123456789" validate:"required"`
}

// Validate проверяет корректность события
func (e *SearchEvent) Validate() error {
	if e.Query == "" {
		return ErrEmptyQuery
	}
	if e.UserID == "" {
		return ErrEmptyUserID
	}
	if e.Timestamp <= 0 {
		return ErrInvalidTimestamp
	}
	return nil
}
