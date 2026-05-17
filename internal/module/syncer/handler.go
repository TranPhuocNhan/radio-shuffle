package syncer

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

func (h *Handler) Trigger(c *gin.Context) {
	userID, ok := mw.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "missing user")
		return
	}
	job, err := h.svc.Trigger(c.Request.Context(), "admin:"+strconv.FormatInt(userID, 10))
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OK(c, http.StatusAccepted, TriggerSyncResponse{
		RequestID: job.RequestID,
		StatusURL: "/api/v1/syncer/status/" + job.RequestID,
	})
}

func (h *Handler) Status(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		response.BadRequest(c, "missing request_id")
		return
	}
	job, err := h.svc.Status(c.Request.Context(), requestID)
	if err != nil {
		if err == ErrJobNotFound {
			response.NotFound(c, "sync job not found")
			return
		}
		response.Internal(c, err)
		return
	}
	response.OK(c, http.StatusOK, toSyncStatusResponse(job))
}


