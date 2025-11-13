package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sssseraphim/Chirpy/internal/auth"
	"github.com/sssseraphim/Chirpy/internal/database"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	Red          bool      `json:"chirpy_red"`
}

type userReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {
	var u userReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&u)
	if err != nil {
		respondWithError(w, 400, "bad email")
		return
	}
	hashedPassword, err := auth.HashPassword(u.Password)
	if err != nil {
		respondWithError(w, 400, fmt.Sprint(err))
		return
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          u.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("couldnt create a user: %v", err))
		return
	}
	respUser := User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email, Red: user.ChirpyRed}
	respondWithJson(w, 201, respUser)
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var u userReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&u)
	if err != nil {
		respondWithError(w, 400, "bad req")
		return
	}
	jwtExpires := time.Hour

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), u.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("no user with such email: %v %v", err, u.Email))
		return
	}
	matches, err := auth.CheckPasswordHash(u.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 500, "we be stupid")
		return
	}
	if !matches {
		respondWithError(w, http.StatusUnauthorized, "wrong password")
		return
	}
	JWToken, err := auth.MakeJWT(user.ID, cfg.secret, jwtExpires)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "failed to make jwt")
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "failed to make refresh token")
		return
	}
	_, err = cfg.dbQueries.AddNewRefreshToken(r.Context(), database.AddNewRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		respondWithError(w, 500, "failed to add refresh token")
		return
	}
	respondWithJson(w, http.StatusOK, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        JWToken,
		RefreshToken: refreshToken,
		Red:          user.ChirpyRed})
}

func (cfg *apiConfig) resetUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, "get the fuck out")
		return
	}
	err := cfg.dbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, 400, "coudnt clear users")
		return
	}
	type devMsg struct {
		Msg string `json:"msg"`
	}
	msg := devMsg{Msg: "clear!"}
	respondWithJson(w, 200, msg)
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "no refresh token given")
		return
	}
	tokenString := token[7:]
	refreshToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, 401, "cant get a token from db")
		return
	}
	if refreshToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, 401, "refresh token expired")
		return
	}
	type payload struct {
		Token string `json:"token"`
	}
	newJwt, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, 500, "failde to make jwt")
		return
	}
	respondWithJson(w, 200, payload{Token: newJwt})
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "no refresh token given")
		return
	}
	tokenString := token[7:]
	_, err := cfg.dbQueries.GetRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, 401, "cant get a token from db")
		return
	}
	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), tokenString)
	if err != nil {
		respondWithError(w, 500, "failed to revoke")
		return
	}
	w.WriteHeader(204)
}

func (cfg *apiConfig) handleChangeEmailAndPassword(w http.ResponseWriter, r *http.Request) {
	jwtoken := r.Header.Get("Authorization")
	if jwtoken == "" {
		respondWithError(w, 401, "no jwt given")
		return
	}
	tokenString := jwtoken[7:]
	userId, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "cant validate jwt")
		return
	}
	user, err := cfg.dbQueries.GetUserByID(r.Context(), userId)
	if err != nil {
		respondWithError(w, 401, "no user found")
		return
	}

	var newUserInfo userReq
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&newUserInfo)
	hashed_password, err := auth.HashPassword(newUserInfo.Password)
	if err != nil {
		respondWithError(w, 500, "failed to hash password")
		return
	}

	user, err = cfg.dbQueries.ChangeUserInfo(r.Context(), database.ChangeUserInfoParams{
		Email:          newUserInfo.Email,
		HashedPassword: hashed_password,
		ID:             user.ID,
	})
	if err != nil {
		respondWithError(w, 500, "failed to change user info")
		return
	}
	respondWithJson(w, 200, User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Red:       user.ChirpyRed,
	})
}

func (cfg *apiConfig) handleWebhooks(w http.ResponseWriter, r *http.Request) {
	type WebhooksData struct {
		User_id uuid.UUID `json:"user_id"`
	}

	type Event struct {
		Event string       `json:"event"`
		Data  WebhooksData `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	var event Event
	err := decoder.Decode(&event)
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	if event.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	api_key, err := auth.GetApiKey(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}
	if api_key != cfg.polkaKey {
		respondWithError(w, 401, "wrong key")
		return
	}
	err = cfg.dbQueries.MakeUserRed(r.Context(), event.Data.User_id)
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}
	w.WriteHeader(204)
}
