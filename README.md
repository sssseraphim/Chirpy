# Chirpy API

A RESTful API service for managing users and chirps (microblog posts) with authentication and webhook support.

## Table of Contents
- [Features](#features)
- [API Endpoints](#api-endpoints)
- [Authentication](#authentication)
- [Webhooks](#webhooks)
- [Admin Endpoints](#admin-endpoints)
- [Setup](#setup)

## Features

- **User Management**: Create users, login, and manage profiles
- **Chirp System**: Create, read, update, and delete chirps (microblog posts)
- **JWT Authentication**: Secure endpoints with JSON Web Tokens
- **Refresh Token Rotation**: Enhanced security with token refresh mechanism
- **Webhook Support**: Integration with external services via webhooks
- **Metrics & Monitoring**: Admin endpoints for monitoring and management
- **Static File Serving**: Serve static assets from the root directory

## API Endpoints

### Health Check
- `GET /api/healthz` - Health check endpoint

### User Management
- `POST /api/users` - Create a new user
- `POST /api/login` - User login
- `PUT /api/users` - Update user email and password
- `POST /api/refresh` - Refresh access token
- `POST /api/revoke` - Revoke refresh token

### Chirp Management
- `POST /api/chirps` - Create a new chirp
- `GET /api/chirps` - Get all chirps
- `GET /api/chirps/{chirpID}` - Get a specific chirp by ID
- `DELETE /api/chirps/{chirpID}` - Delete a chirp

### Webhooks
- `POST /api/polka/webhooks` - Handle Polka webhook events

## Authentication

This API uses JWT (JSON Web Tokens) for authentication:

1. **Login**: Use `POST /api/login` to get access and refresh tokens
2. **Access Token**: Short-lived token for API access (sent in Authorization header)
3. **Refresh Token**: Long-lived token for obtaining new access tokens
4. **Token Refresh**: Use `POST /api/refresh` to get a new access token
5. **Token Revocation**: Use `POST /api/revoke` to invalidate a refresh token

## Webhooks

The API supports webhook integrations:

- `POST /api/polka/webhooks` - Handle events from Polka service

## Admin Endpoints

Administrative endpoints for monitoring and management:

- `GET /admin/metrics` - View application metrics and hit counts
- `POST /admin/reset` - Reset user data (development use)

### Static Files

- `GET /app/*` - Serve static files from the application root directory

## Setup

### Prerequisites
- Go 1.21+
- PostgreSQL database

### Configuration
```go
cfg.dbQueries = database.New(db)
const filepathRoot = "."
const port = ":8080"
```

### Running the Server
```bash
go run main.go
```

The server will start on port 8080 and serve:
- API endpoints under `/api/`
- Admin endpoints under `/admin/`
- Static files under `/app/`

### Environment Variables
- Database connection string
- JWT secret key
- Polka webhook API key

---

For more information, please refer to the API documentation or contact the development team.
