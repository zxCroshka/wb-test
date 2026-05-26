package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleHealth(c *gin.Context) {
	// Отладочный вывод в консоль Docker
	println("=== HEALTH HANDLER CALLED ===")

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
func (s *Server) handleTop(c *gin.Context) {
	var req TopRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("invalid limit parameter", 400, err.Error()))
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	top, err := s.aggregator.GetTop(c, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("failed to get top", 500, err.Error()))
		return
	}

	items := ToRankItemResponse(top)

	c.JSON(http.StatusOK, NewTopResponse(items, time.Now()))
}

func (s *Server) handleStopListAdd(c *gin.Context) {
	var req StopListAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("invalid request", 400, err.Error()))
		return
	}

	s.aggregator.AddToStopList(req.Word)

	c.JSON(http.StatusOK, StopListAddResponse{
		Success: true,
		Message: "Word added to stoplist",
		Word:    req.Word,
	})
}

func (s *Server) handleStopListRemove(c *gin.Context) {
	var req StopListRemoveRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("invalid request", 400, err.Error()))
		return
	}

	if removed := s.aggregator.RemoveFromStopList(req.Word); removed {
		c.JSON(http.StatusOK, StopListRemoveResponse{
			Success: true,
			Message: "Word removed from stoplist",
		})
	} else {
		c.JSON(http.StatusNotFound, NewErrorResponse("word not found", 404, ""))
	}
}

func (s *Server) handleStopListGetAll(c *gin.Context) {
	words := s.aggregator.GetStopList()

	c.JSON(http.StatusOK, StopListGetAllResponse{
		Words: words,
		Count: len(words),
	})
}

func (s *Server) handleStopListCheck(c *gin.Context) {
	var req StopListCheckRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("invalid request", 400, err.Error()))
		return
	}

	blocked := s.aggregator.IsWordBlocked(req.Word)

	c.JSON(http.StatusOK, StopListCheckResponse{
		Word:    req.Word,
		Blocked: blocked,
	})
}

func (s *Server) handleStopListBatchAdd(c *gin.Context) {
	var req StopListBatchAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("invalid request", 400, err.Error()))
		return
	}

	added := s.aggregator.BatchAddToStopList(req.Words)

	c.JSON(http.StatusOK, StopListBatchAddResponse{
		Success: true,
		Added:   added,
		Total:   len(req.Words),
	})
}

func (s *Server) handleStopListClear(c *gin.Context) {
	s.aggregator.ClearStopList()

	c.JSON(http.StatusOK, StopListClearResponse{
		Success: true,
		Message: "Stoplist cleared",
	})
}
