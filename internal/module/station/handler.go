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
	station, err := h.svc.Create(c.Request.Context(), toCreateStationInput(req))
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OK(c, http.StatusCreated, toStationResponse(station))
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
	response.OK(c, http.StatusOK, toStationResponse(out))
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
	items, total, err := h.svc.List(c.Request.Context(), ListStationsInput{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		response.Internal(c, err)
		return
	}
	responses := make([]StationResponse, 0, len(items))
	for _, station := range items {
		responses = append(responses, toStationResponse(station))
	}
	page := 1
	if limit > 0 {
		page = int(offset/limit) + 1
	}
	response.Paginated(c, http.StatusOK, responses, page, int(limit), total)
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
	out, err := h.svc.Update(c.Request.Context(), toUpdateStationInput(id, req))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(c, "station not found")
			return
		}
		response.Internal(c, err)
		return
	}
	response.OK(c, http.StatusOK, toStationResponse(out))
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
	return strconv.ParseInt(c.Param("station_id"), 10, 64)
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
