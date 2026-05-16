package radiobrowser

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// Handler exposes radio browser station HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *gin.Context) {
	name := strings.TrimSpace(c.Query("name"))
	country := strings.TrimSpace(c.Query("country"))
	language := strings.TrimSpace(c.Query("language"))
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
		Name:     name,
		Country:  country,
		Language: language,
		Limit:    limit,
		Offset:   offset,
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
