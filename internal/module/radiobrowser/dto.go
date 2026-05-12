package radiobrowser

import "time"

// StationResponse is the JSON shape returned for a radio browser station.
type StationResponse struct {
	ID          int64     `json:"id"`
	StationUUID string    `json:"station_uuid"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	URLResolved *string   `json:"url_resolved"`
	Homepage    *string   `json:"homepage"`
	Favicon     *string   `json:"favicon"`
	Country     *string   `json:"country"`
	CountryCode *string   `json:"country_code"`
	State       *string   `json:"state"`
	Language    *string   `json:"language"`
	Codec       *string   `json:"codec"`
	Bitrate     int32     `json:"bitrate"`
	Votes       int32     `json:"votes"`
	Tags        *string   `json:"tags"`
	LastCheckOK bool      `json:"last_check_ok"`
	SyncedAt    time.Time `json:"synced_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toStationResponse(r Station) StationResponse {
	return StationResponse{
		ID:          r.ID,
		StationUUID: r.StationUUID,
		Name:        r.Name,
		URL:         r.URL,
		URLResolved: r.URLResolved,
		Homepage:    r.Homepage,
		Favicon:     r.Favicon,
		Country:     r.Country,
		CountryCode: r.CountryCode,
		State:       r.State,
		Language:    r.Language,
		Codec:       r.Codec,
		Bitrate:     r.Bitrate,
		Votes:       r.Votes,
		Tags:        r.Tags,
		LastCheckOK: r.LastCheckOK,
		SyncedAt:    r.SyncedAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
