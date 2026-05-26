package dto

type SearchEvent struct {
	Query     string `json:"query"`
	UserID    string `json:"user_id"`
	Timestamp int64  `json:"timestamp"`
}

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
