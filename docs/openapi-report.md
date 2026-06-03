# OpenAPI Implementation Report

Generated spec: `docs/openapi.yaml`

## Endpoint Inventory

| Method | Path | Handler | Auth required | Request DTO | Success response DTO |
|---|---|---|---|---|---|
| GET | `/swagger` | `apidocs.SwaggerUI` | No | None | Redirect to Swagger UI |
| GET | `/swagger/index.html` | `gin-swagger` library handler | No | None | HTML |
| GET | `/openapi.yaml` | `apidocs.OpenAPIYAML` | No | None | OpenAPI YAML |
| GET | `/health` | `health.Handler.Live` | No | None | Plain text |
| GET | `/ready` | `health.Handler.Ready` | No | None | Plain text |
| GET | `/api/v1/ping` | Inline router handler | No | None | Inline `{service}` envelope |
| POST | `/api/v1/auth/register` | `auth.Handler.Register` | No | `RegisterRequest` | `AuthResponse` |
| POST | `/api/v1/auth/login` | `auth.Handler.Login` | No | `LoginRequest` | `AuthResponse` |
| POST | `/api/v1/auth/refresh` | `auth.Handler.Refresh` | No | `RefreshRequest` | `AuthResponse` |
| POST | `/api/v1/auth/logout` | `auth.Handler.Logout` | No | `LogoutRequest` | No content |
| POST | `/api/v1/stations` | `station.Handler.Create` | No | `CreateStationRequest` | `StationResponse` |
| GET | `/api/v1/stations` | `station.Handler.List` | No | None | Paginated `StationResponse[]` |
| GET | `/api/v1/stations/followed` | `station.Handler.ListFollowed` | Bearer JWT | None | Paginated `StationResponse[]` |
| GET | `/api/v1/stations/{station_id}` | `station.Handler.GetByID` | No | None | `StationResponse` |
| PATCH | `/api/v1/stations/{station_id}` | `station.Handler.Update` | No | `UpdateStationRequest` | `StationResponse` |
| DELETE | `/api/v1/stations/{station_id}` | `station.Handler.Delete` | No | None | No content |
| POST | `/api/v1/stations/{station_id}/follow` | `station.Handler.Follow` | Bearer JWT | None | No content |
| DELETE | `/api/v1/stations/{station_id}/follow` | `station.Handler.Unfollow` | Bearer JWT | None | No content |
| GET | `/api/v1/stations/{station_id}/following` | `station.Handler.IsFollowing` | Bearer JWT | None | `FollowStatusResponse` |
| GET | `/api/v1/stations/{station_id}/followers/count` | `station.Handler.CountFollowers` | Bearer JWT | None | `FollowersCountResponse` |
| GET | `/api/v1/stations/{station_id}/tracks` | `track.Handler.List` | No | None | Paginated `TrackResponse[]` |
| POST | `/api/v1/stations/{station_id}/tracks` | `track.Handler.Create` | Bearer JWT | `CreateTrackRequest` | `TrackResponse` |
| GET | `/api/v1/stations/{station_id}/tracks/{id}` | `track.Handler.GetByID` | No | None | `TrackResponse` |
| PATCH | `/api/v1/stations/{station_id}/tracks/{id}` | `track.Handler.Update` | Bearer JWT | `UpdateTrackRequestBody` | `TrackResponse` |
| DELETE | `/api/v1/stations/{station_id}/tracks/{id}` | `track.Handler.Delete` | Bearer JWT | None | No content |
| POST | `/api/v1/playlists` | `playlist.Handler.create` | Bearer JWT | `CreatePlaylistRequest` | `PlaylistResponse` |
| GET | `/api/v1/playlists` | `playlist.Handler.list` | Bearer JWT | None | Paginated `PlaylistResponse[]` |
| GET | `/api/v1/playlists/{id}` | `playlist.Handler.getByID` | Bearer JWT | None | `PlaylistResponse` |
| PATCH | `/api/v1/playlists/{id}` | `playlist.Handler.update` | Bearer JWT | `UpdatePlaylistRequest` | `PlaylistResponse` |
| DELETE | `/api/v1/playlists/{id}` | `playlist.Handler.delete` | Bearer JWT | None | No content |
| POST | `/api/v1/playlists/{id}/tracks` | `playlist.Handler.addTrack` | Bearer JWT | `AddPlaylistTrackRequest` | No content |
| GET | `/api/v1/playlists/{id}/tracks` | `playlist.Handler.listTracks` | Bearer JWT | None | Paginated `PlaylistTrackResponse[]` |
| DELETE | `/api/v1/playlists/{id}/tracks/{track_id}` | `playlist.Handler.removeTrack` | Bearer JWT | None | No content |
| PATCH | `/api/v1/playlists/{id}/tracks/reorder` | `playlist.Handler.reorderTracks` | Bearer JWT | `ReorderPlaylistTracksRequest` | No content |
| POST | `/api/v1/streams` | `stream.Handler.Start` | Bearer JWT | `StartStreamRequest` | `StreamResponse` |
| GET | `/api/v1/streams` | `stream.Handler.List` | Bearer JWT | None | Paginated `StreamResponse[]` |
| GET | `/api/v1/streams/{id}` | `stream.Handler.GetByID` | Bearer JWT | None | `StreamResponse` |
| PATCH | `/api/v1/streams/{id}/end` | `stream.Handler.End` | Bearer JWT | None | `StreamResponse` |
| GET | `/api/v1/radio-browser/stations` | `radiobrowser.Handler.List` | No | None | Paginated `RadioBrowserStationResponse[]` |
| POST | `/api/v1/syncer/trigger` | `syncer.Handler.Trigger` | Bearer JWT + admin role | None | `TriggerSyncResponse` |
| GET | `/api/v1/syncer/status/{request_id}` | `syncer.Handler.Status` | Bearer JWT + admin role | None | `SyncStatusResponse` |

## Undocumented Endpoints

None found in the route registrations scanned from `cmd/api/main.go`, `internal/router/router.go`, and `internal/module/*/module.go`.

## Ambiguous Endpoints

- `GET /api/v1/stations/followed` overlaps structurally with `GET /api/v1/stations/{station_id}`. OpenAPI can document both, but the Gin route tree should be covered by an API-level router test because the static route is registered after the parameterized `GET /:station_id`.
- `GET /api/v1/stations/{station_id}/followers/count` is wired behind auth middleware, but the handler does not use the authenticated user. The spec documents the actual route requirement.

## Missing DTOs

- `GET /api/v1/ping` returns an inline map instead of a named DTO.
- `/health` and `/ready` intentionally return `text/plain`, not the shared JSON envelope.
- No request DTO exists for endpoints with no request body. The OpenAPI spec omits `requestBody` for those operations.

## Schema Comparison Notes

- The OpenAPI component schemas mirror `json` tags and `binding:"required"` tags from the handler DTOs.
- Pagination uses the actual handler behavior: `limit` defaults to 20, values above 100 are capped, and negative `offset` values are normalized to 0 after parsing.
- `UpdatePlaylistRequest` documents the handler-level "at least one field is required" rule with `minProperties: 1`.
- `ReorderPlaylistTracksRequest.items` documents the handler-level non-empty rule with `minItems: 1`.
- Error responses use the shared `ErrorEnvelope`; health readiness errors remain plain text because `health.Handler.Ready` writes `c.String`.

## Implementation Mismatches And Risks

- Station create, update, and delete are public routes, but `CreateStationRequest` and `UpdateStationRequest` allow `owner_id`. If station ownership is intended, these routes should probably require auth and derive `owner_id` from context.
- Track `GET` and `DELETE` parse `station_id` but the service methods operate only by track ID. A request under one station can address a track that belongs to another station unless the repository/service enforces the station relation.
- Track `PATCH` reads the track by ID first and then updates using the path `station_id`; this can behave like a station move if the target ID belongs to a different station.
- `UpdateStationRequest` accepts an empty JSON object and performs a read-then-merge update with no changed fields. Other modules reject empty PATCH payloads.
- `SyncStatusResponse.scope` is `json.RawMessage`; the spec can only document it as arbitrary raw JSON until the sync scope becomes a typed DTO.

## Suggested Improvements

- Add a router integration test that builds the real `cmd/api` route tree or a lightweight equivalent and asserts all registered paths, especially `/stations/followed`.
- Introduce named DTOs for `PingResponse` and, if desired, health responses to reduce inline shapes.
- Add OpenAPI validation to CI, for example with `redocly lint` or `swagger-cli validate`.
- Consider adding auth and owner derivation for station writes if station ownership should be enforced.
- Enforce station scoping in track service/repository methods for `GET`, `PATCH`, and `DELETE`.
