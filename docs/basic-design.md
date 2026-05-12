# Basic Design

## Purpose
Radio Shuffle is a backend for a radio and playlist service. It provides authentication, station management, track management, and a background syncer that ingests public stations from the Radio Browser dataset.

## Business Terms
- User: A person using the service. Users have roles (admin or user).
- Station: A radio station owned by a user or public to all listeners.
- Track: A playable audio item associated with a station.
- Playlist: A curated list of tracks owned by a user.
- Stream: A listening session for a user and station, with start and end times.
- Refresh Token: A long-lived token for issuing new access tokens.
- Radio Browser Station: A station record ingested from the external Radio Browser source.

## Actors
- Listener: Browses and listens to public stations and tracks.
- Creator: Creates and manages stations, tracks, and playlists.
- Admin: Manages users and oversees content (future scope).
- Syncer Service: Background process ingesting external station data.

## High-Level Architecture
- Modular monolith with domain modules under `internal/module/`.
- Handler -> Service -> Repository layering for all business features.
- SQL-first data access via SQLC with PostgreSQL as the primary database.
- Standardized response envelope for all HTTP responses.

## Core Flows
1. User registration and login
   - User registers with email and password.
   - Service issues access and refresh tokens.
   - User refreshes or logs out via refresh token.
2. Discover and browse stations
   - List public stations with pagination.
   - View station details.
3. Manage stations
   - Create station with name, stream URL, and optional metadata.
   - Update station fields; partial updates preserve prior values.
   - Delete station.
4. Manage tracks
   - Create, update, list, and delete tracks scoped to a station.
   - Read-only access is public; write access requires authentication.
5. Playlist management (planned)
   - Create playlists and add tracks.
   - Manage ordering of tracks within a playlist.
6. Streaming sessions (planned)
   - Record a stream session when a user plays a station.
7. External station ingestion
   - Syncer fetches stations from Radio Browser on a schedule.
   - Upserts ingest data into the local `radio_browser_stations` table.

## Data Model Summary
- Users own stations and playlists; playlists and stations can be public.
- Tracks belong to exactly one station.
- Playlist and track relationship is many-to-many with explicit ordering.
- Streams capture user listening sessions by station.

## Status Summary
- Complete: auth, station, track, syncer.
- Repository-only: user.
- Stubbed: playlist, stream.

## Out of Scope (Current)
- Full playlist and stream HTTP APIs.
- Admin tooling and role-based management.
- Recommendation or analytics features.

