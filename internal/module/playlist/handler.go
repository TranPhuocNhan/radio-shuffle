package playlist

import (
	"net/http"
	"strconv"

	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"

	"github.com/gin-gonic/gin"
)

// Handler exposes playlist HTTP endpoints.
type Handler struct {
	svc Service
}

// NewHandler constructs a Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) create(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	var req CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	row, err := h.svc.Create(c.Request.Context(), userID, CreatePlaylistInput{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		return err
	}
	response.OK(c, http.StatusCreated, toPlaylistResponse(row))
	return nil
}

func (h *Handler) list(c *gin.Context) error {
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
	responses := make([]PlaylistResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toPlaylistResponse(item))
	}
	page := 1
	if limit > 0 {
		page = int(offset/limit) + 1
	}
	response.Paginated(c, http.StatusOK, responses, page, int(limit), total)
	return nil
}

func (h *Handler) getByID(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	id, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	out, err := h.svc.GetByID(c.Request.Context(), id, userID)
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, toPlaylistResponse(out))
	return nil
}

func (h *Handler) update(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	id, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	var req UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	if req.Name == nil && req.Description == nil && req.IsPublic == nil {
		return httperr.BadRequest("at least one field is required")
	}
	out, err := h.svc.Update(c.Request.Context(), id, userID, UpdatePlaylistInput{
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, toPlaylistResponse(out))
	return nil
}

func (h *Handler) delete(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	id, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	if err := h.svc.Delete(c.Request.Context(), id, userID); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) addTrack(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	playlistID, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	var req AddPlaylistTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	if err := h.svc.AddTrack(c.Request.Context(), playlistID, userID, req.TrackID, *req.Position); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) removeTrack(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	playlistID, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	trackID, err := parseTrackIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid track_id")
	}
	if err := h.svc.RemoveTrack(c.Request.Context(), playlistID, userID, trackID); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) listTracks(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	playlistID, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
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
	items, total, err := h.svc.ListTracks(c.Request.Context(), playlistID, userID, limit, offset)
	if err != nil {
		return err
	}
	responses := make([]PlaylistTrackResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, toPlaylistTrackResponse(item))
	}
	page := 1
	if limit > 0 {
		page = int(offset/limit) + 1
	}
	response.Paginated(c, http.StatusOK, responses, page, int(limit), total)
	return nil
}

func (h *Handler) reorderTracks(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	playlistID, err := parseIDParam(c)
	if err != nil {
		return httperr.BadRequest("invalid id")
	}
	var req ReorderPlaylistTracksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return httperr.BadRequest(err.Error())
	}
	if len(req.Items) == 0 {
		return httperr.BadRequest("items cannot be empty")
	}
	patches := make([]TrackPositionPatch, 0, len(req.Items))
	for _, item := range req.Items {
		patches = append(patches, TrackPositionPatch{TrackID: item.TrackID, Position: item.Position})
	}
	if err := h.svc.ReorderTracks(c.Request.Context(), playlistID, userID, patches); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func parseIDParam(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func parseTrackIDParam(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("track_id"), 10, 64)
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
