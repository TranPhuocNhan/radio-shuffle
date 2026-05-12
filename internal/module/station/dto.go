package station

import "time"

// CreateStationRequest is the HTTP body for POST /stations.
type CreateStationRequest struct {
	Name          string  `json:"name" binding:"required"`
	Genre         string  `json:"genre"`
	Description   *string `json:"description"`
	StreamUrl     string  `json:"stream_url" binding:"required"`
	CoverImageUrl *string `json:"cover_image_url"`
	IsPublic      *bool   `json:"is_public,omitempty"`
	OwnerID       *int64  `json:"owner_id,omitempty"`
}

// UpdateStationRequest is the HTTP body for PATCH /stations/:station_id.
// Omitted fields keep the existing values.
type UpdateStationRequest struct {
	Name          *string `json:"name"`
	Genre         *string `json:"genre"`
	Description   *string `json:"description"`
	StreamUrl     *string `json:"stream_url"`
	CoverImageUrl *string `json:"cover_image_url"`
	IsPublic      *bool   `json:"is_public"`
	OwnerID       *int64  `json:"owner_id"`
}

// StationResponse is the JSON shape returned for a single station.
type StationResponse struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Genre         string    `json:"genre"`
	Description   *string   `json:"description"`
	StreamUrl     string    `json:"stream_url"`
	CoverImageUrl *string   `json:"cover_image_url"`
	IsPublic      bool      `json:"is_public"`
	OwnerID       *int64    `json:"owner_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func toCreateStationInput(req CreateStationRequest) CreateStationInput {
	return CreateStationInput{
		Name:          req.Name,
		Genre:         req.Genre,
		Description:   req.Description,
		StreamUrl:     req.StreamUrl,
		CoverImageUrl: req.CoverImageUrl,
		IsPublic:      req.IsPublic,
		OwnerID:       req.OwnerID,
	}
}

func toUpdateStationInput(id int64, req UpdateStationRequest) UpdateStationInput {
	return UpdateStationInput{
		ID:            id,
		Name:          req.Name,
		Genre:         req.Genre,
		Description:   req.Description,
		StreamUrl:     req.StreamUrl,
		CoverImageUrl: req.CoverImageUrl,
		IsPublic:      req.IsPublic,
		OwnerID:       req.OwnerID,
	}
}

func toStationResponse(r Station) StationResponse {
	return StationResponse{
		ID:            r.ID,
		Name:          r.Name,
		Genre:         r.Genre,
		Description:   r.Description,
		StreamUrl:     r.StreamUrl,
		CoverImageUrl: r.CoverImageUrl,
		IsPublic:      r.IsPublic,
		OwnerID:       r.OwnerID,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
