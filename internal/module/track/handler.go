package track

import (
	"net/http"
	"strconv"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// Handler exposes track HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) error {
	stationID, err := parseStationIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid station_id")
	}
	var req CreateTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	row, err := h.svc.Create(c.Request.Context(), CreateTrackInput{
		StationID:       stationID,
		Title:           req.Title,
		Artist:          req.Artist,
		AudioUrl:        req.AudioUrl,
		DurationSeconds: req.DurationSeconds,
	})
	if err != nil {
		return err
	}
	response.OK(c, http.StatusCreated, toTrackResponse(row))
	return nil
}

func (h *Handler) GetByID(c *gin.Context) error {
	_, err := parseStationIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid station_id")
	}
	id, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, toTrackResponse(out))
	return nil
}

func (h *Handler) List(c *gin.Context) error {
	stationID, err := parseStationIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid station_id")
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
	tracks, total, err := h.svc.List(c.Request.Context(), stationID, limit, offset)
	if err != nil {
		return err
	}
	items := make([]TrackResponse, 0, len(tracks))
	for _, t := range tracks {
		items = append(items, toTrackResponse(t))
	}
	page := 1
	if limit > 0 {
		page = int(offset/limit) + 1
	}
	response.Paginated(c, http.StatusOK, items, page, int(limit), total)
	return nil
}

func (h *Handler) Update(c *gin.Context) error {
	stationID, err := parseStationIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid station_id")
	}
	id, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	var req UpdateTrackRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	out, err := h.svc.Update(c.Request.Context(), id, stationID, UpdateTrackInput{
		Title:           req.Title,
		Artist:          req.Artist,
		AudioUrl:        req.AudioUrl,
		DurationSeconds: req.DurationSeconds,
	})
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, toTrackResponse(out))
	return nil
}

func (h *Handler) Delete(c *gin.Context) error {
	_, err := parseStationIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid station_id")
	}
	id, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func parseStationIDParam(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("station_id"), 10, 64)
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
