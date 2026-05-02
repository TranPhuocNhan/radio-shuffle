package station

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// Handler exposes station HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	row, err := h.svc.Create(c.Request.Context(), CreateStationInput{
		Name:          req.Name,
		Genre:         req.Genre,
		Description:   req.Description,
		StreamUrl:     req.StreamUrl,
		CoverImageUrl: req.CoverImageUrl,
		IsPublic:      req.IsPublic,
		OwnerID:       req.OwnerID,
	})
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OK(c, http.StatusCreated, toStationResponse(row))
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "station not found")
			return
		}
		response.Internal(c, err)
		return
	}
	response.OK(c, http.StatusOK, out)
}

func (h *Handler) List(c *gin.Context) {
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
	items, total, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.Internal(c, err)
		return
	}
	page := 1
	if limit > 0 {
		page = int(offset/limit) + 1
	}
	response.Paginated(c, http.StatusOK, items, page, int(limit), total)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var req UpdateStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "station not found")
			return
		}
		response.Internal(c, err)
		return
	}
	response.OK(c, http.StatusOK, out)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := parseIDParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "station not found")
			return
		}
		response.Internal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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
