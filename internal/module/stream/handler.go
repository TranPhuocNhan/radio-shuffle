package stream

import (
	"net/http"
	"strconv"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
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

func (h *Handler) Start(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	var req StartStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	row, err := h.svc.Start(c.Request.Context(), userID, req.StationID)
	if err != nil {
		return err
	}
	response.OK(c, http.StatusCreated, toStreamResponse(row))
	return nil
}

func (h *Handler) GetByID(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	id, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	row, err := h.svc.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, toStreamResponse(row))
	return nil
}

func (h *Handler) List(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	limit, err := parseQueryInt64(c, "limit", defaultListLimit)
	if err != nil {
		return httperr.BadRequest("invalid limit")
	}
	offset, err := parseQueryInt64(c, "offset", 0)
	if err != nil {
		return httperr.BadRequest("invalid offset")
	}
	limit, offset = normalizeListParams(limit, offset)
	items, total, err := h.svc.List(c.Request.Context(), userID, limit, offset)
	if err != nil {
		return err
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
	return nil
}

func (h *Handler) End(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	id, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	row, err := h.svc.End(c.Request.Context(), id, userID)
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, toStreamResponse(row))
	return nil
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
