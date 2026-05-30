package syncer

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/httperr"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/mw"
	"github.com/tranphuocnhan/radio-shuffle/internal/platform/response"
)

// Handler exposes syncer admin HTTP endpoints.
type Handler struct {
	svc JobService
}

// NewHandler constructs a Handler.
func NewHandler(svc JobService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Trigger(c *gin.Context) error {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		return httperr.Unauthorized("missing user")
	}
	job, err := h.svc.Trigger(c.Request.Context(), "admin:"+strconv.FormatInt(userID, 10))
	if err != nil {
		return err
	}
	response.OK(c, http.StatusAccepted, TriggerSyncResponse{
		RequestID: job.RequestID,
		StatusURL: "/api/v1/syncer/status/" + job.RequestID,
	})
	return nil
}

func (h *Handler) Status(c *gin.Context) error {
	requestID := c.Param("request_id")
	if requestID == "" {
		return httperr.BadRequest("missing request_id")
	}
	job, err := h.svc.Status(c.Request.Context(), requestID)
	if err != nil {
		return err
	}
	response.OK(c, http.StatusOK, toSyncStatusResponse(job))
	return nil
}
