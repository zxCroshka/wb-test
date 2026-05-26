package api

import (
	"time"

	"github.com/zxCroshka/wb-test/internal/domain"
)

type TopRequest struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

type StopListAddRequest struct {
	Word string `json:"word" binding:"required,min=1,max=100"`
}

type StopListRemoveRequest struct {
	Word string `form:"word" binding:"required,min=1,max=100"`
}

type StopListBatchAddRequest struct {
	Words []string `json:"words" binding:"required,min=1,max=1000,dive,min=1,max=100"`
}

type StopListCheckRequest struct {
	Word string `form:"word" binding:"required,min=1,max=100"`
}

type RankItemResponse struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}

type TopResponse struct {
	Top       []RankItemResponse `json:"top"`
	Total     int                `json:"total"`
	UpdatedAt string             `json:"updated_at,omitempty"`
}

type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

type StopListAddResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Word    string `json:"word"`
}

type StopListRemoveResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type StopListGetAllResponse struct {
	Words []string `json:"words"`
	Count int      `json:"count"`
}

type StopListCheckResponse struct {
	Word    string `json:"word"`
	Blocked bool   `json:"blocked"`
}

type StopListBatchAddResponse struct {
	Success bool `json:"success"`
	Added   int  `json:"added"`
	Total   int  `json:"total"`
}

type StopListClearResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

func NewTopResponse(items []RankItemResponse, updatedAt time.Time) TopResponse {
	return TopResponse{
		Top:       items,
		Total:     len(items),
		UpdatedAt: updatedAt.UTC().Format(time.RFC3339),
	}
}

func NewHealthResponse() HealthResponse {
	return HealthResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	}
}

func NewErrorResponse(err string, code int, details string) ErrorResponse {
	return ErrorResponse{
		Error:   err,
		Code:    code,
		Details: details,
	}
}

func ToRankItemResponse(items []domain.RankItem) []RankItemResponse {
	if items == nil {
		return []RankItemResponse{}
	}

	result := make([]RankItemResponse, len(items))
	for i, item := range items {
		result[i] = RankItemResponse{
			Query: item.Query,
			Count: item.Count,
		}
	}
	return result
}
