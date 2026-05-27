package stream

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// Handler exposes stream HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Start(c *gin.Context) {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	var req StartStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	row, err := h.svc.Start(c.Request.Context(), userID, req.StationID)
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OK(c, http.StatusCreated, toStreamResponse(row))
}

func (h *Handler) GetByID(c *gin.Context) {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	row, err := h.svc.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			response.NotFound(c, "stream not found")
		case errors.Is(err, ErrForbidden):
			response.Forbidden(c, "forbidden")
		default:
			response.Internal(c, err)
		}
		return
	}
	response.OK(c, http.StatusOK, toStreamResponse(row))
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	limit, err := parseQueryInt64(c, "limit", defaultListLimit)
	if err != nil {
		response.BadRequest(c, "invalid limit")
		return
	}
	offset, err := parseQueryInt64(c, "offset", 0)
	if err != nil {
		response.BadRequest(c, "invalid offset")
		return
	}
	limit, offset = normalizeListParams(limit, offset)
	items, total, err := h.svc.List(c.Request.Context(), userID, limit, offset)
	if err != nil {
		response.Internal(c, err)
		return
	}
	responses := make([]StreamResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toStreamResponse(item))
	}
	page := 1
	if limit > 0 {
		page = int(offset/limit) + 1
	}
	response.Paginated(c, http.StatusOK, responses, page, int(limit), total)
}

func (h *Handler) End(c *gin.Context) {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	row, err := h.svc.End(c.Request.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			response.NotFound(c, "stream not found")
		case errors.Is(err, ErrForbidden):
			response.Forbidden(c, "forbidden")
		case errors.Is(err, ErrAlreadyEnded):
			response.Conflict(c, "stream already ended")
		default:
			response.Internal(c, err)
		}
		return
	}
	response.OK(c, http.StatusOK, toStreamResponse(row))
}

func parseIDParam(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func parseQueryInt64(c *gin.Context, key string, defaultVal int64) (int64, error) {
	s := c.Query(key)
	if s == "" {
		return defaultVal, nil
	}
	return strconv.ParseInt(s, 10, 64)
}

func normalizeListParams(limit, offset int64) (int64, int64) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
