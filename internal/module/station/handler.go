package station

import (
	"net/http"
	"strconv"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
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

func (h *Handler) Create(c *gin.Context) error {
	var req CreateStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	station, err := h.svc.Create(c.Request.Context(), toCreateStationInput(req))
	if err != nil {
		return err
	}
	response.OK(c, http.StatusCreated, toStationResponse(station))
	return nil
}

func (h *Handler) GetByID(c *gin.Context) error {
	id, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, toStationResponse(out))
	return nil
}

func (h *Handler) List(c *gin.Context) error {
	limit, err := parseQueryInt64(c, "limit", defaultListLimit)
	if err != nil {
		return httperr.BadRequest("invalid limit")
	}
	offset, err := parseQueryInt64(c, "offset", 0)
	if err != nil {
		return httperr.BadRequest("invalid offset")
	}
	limit, offset = normalizeListParams(limit, offset)
	items, total, err := h.svc.List(c.Request.Context(), ListStationsInput{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return err
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
	return nil
}

func (h *Handler) Update(c *gin.Context) error {
	id, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	var req UpdateStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	out, err := h.svc.Update(c.Request.Context(), toUpdateStationInput(id, req))
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, toStationResponse(out))
	return nil
}

func (h *Handler) Delete(c *gin.Context) error {
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

func (h *Handler) Follow(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	stationID, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	if err := h.svc.Follow(c.Request.Context(), userID, stationID); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) Unfollow(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	stationID, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	if err := h.svc.Unfollow(c.Request.Context(), userID, stationID); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) IsFollowing(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	stationID, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	following, err := h.svc.IsFollowing(c.Request.Context(), userID, stationID)
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, FollowStatusResponse{Following: following})
	return nil
}

func (h *Handler) ListFollowed(c *gin.Context) error {
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
	total, err := h.svc.CountFollowedStations(c.Request.Context(), userID)
	if err != nil {
		return err
	}
	items, err := h.svc.GetFollowedStations(c.Request.Context(), userID, limit, offset)
	if err != nil {
		return err
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
	return nil
}

func (h *Handler) CountFollowers(c *gin.Context) error {
	stationID, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	count, err := h.svc.CountFollowers(c.Request.Context(), stationID)
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, FollowersCountResponse{Count: count})
	return nil
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
