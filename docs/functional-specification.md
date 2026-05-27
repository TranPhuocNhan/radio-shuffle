# Functional Specification (Business-Level)

## Scope
This specification covers the business use cases for Radio Shuffle. It stays at the business level and avoids endpoint or implementation details. Some modules are complete, others are planned.

## Actors
- Listener: Reads and listens to public content.
- Creator: Manages stations, tracks, and playlists.
- Admin: Oversees platform and users (future scope).
- Syncer Service: Periodically imports external stations.

## User and Auth
### Use Cases
- Register a new user account.
- Log in to receive access and refresh tokens.
- Refresh an expired access token using a refresh token.
- Log out by invalidating a refresh token.

### Business Rules
- Email must be unique per user.
- Password must meet minimum strength requirements.
- Access tokens are short-lived; refresh tokens are longer-lived.
- Refresh tokens are stored as hashes and invalidated on refresh or logout.

## Stations
### Use Cases
- Create a station with a name and stream URL.
- List stations with pagination.
- View a station by its identifier.
- Update station metadata.
- Delete a station.

### Business Rules
- Stations can be public or private.
- A station may have an owner (user) but can also be unowned.
- Partial updates keep unspecified fields unchanged.
- Deleting a station removes all associated tracks.

## Tracks
### Use Cases
- Add tracks to a station.
- List tracks for a station with pagination.
- View a track by its identifier.
- Update track metadata.
- Remove tracks from a station.

### Business Rules
- Tracks always belong to one station.
- Track management operations require authentication (read is public).
- Partial updates keep unspecified fields unchanged.

## Playlists
### Use Cases
- Create a playlist owned by a user.
- Add tracks to a playlist.
- Reorder tracks in a playlist.
- Share a playlist publicly or keep it private.

### Business Rules
- Playlists are owned by a user.
- Playlist visibility can be public or private.
- A track can appear in multiple playlists.

## Streams
### Use Cases
- Record when a user starts listening to a station.
- Record when a user ends a listening session.
- View a user stream history (future scope).

### Business Rules
- A stream references both the user and station.
- Ended streams record an end timestamp.

## Radio Browser Sync
### Use Cases
- Periodically ingest stations from the external Radio Browser API.
- Keep the local station index up to date with external data.

### Business Rules
- Ingestion runs on a configurable interval.
- External stations are upserted to avoid duplication.

## Non-Functional Requirements
- Consistent response envelope for all API responses.
- Modular boundaries between domains to avoid tight coupling.
- SQL-driven data access with generated queries for safety.
- Pagination support for large lists.
- Authentication for write operations where applicable.

## Known Gaps and Future Work
- Add admin use cases and role-based management.
- Expand user profile features beyond authentication.

