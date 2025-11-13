package main

import (
	"encoding/json"
	"net/http"
	"slices"

	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sssseraphim/Chirpy/internal/auth"
	"github.com/sssseraphim/Chirpy/internal/database"
)

type ChirpJson struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleChirps(w http.ResponseWriter, r *http.Request) {
	type jsonPayload struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	var pl jsonPayload
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&pl)
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "cant get token")
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "cant validate")
		return
	}
	if userId != pl.UserId {
		respondWithError(w, http.StatusUnauthorized, "wrong id")
		return
	}
	if len(pl.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	pl.Body = cleanBody(pl.Body)

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: pl.Body, UserID: pl.UserId})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("cant create a chirp: %v", err))
		return
	}
	cj := ChirpJson{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}
	respondWithJson(w, 201, cj)
}

func (cfg *apiConfig) handleListChirps(w http.ResponseWriter, r *http.Request) {
	author_id := r.URL.Query().Get("author_id")
	sort_param := r.URL.Query().Get("sort")
	var chirps []database.Chirp
	var err error
	if author_id == "" {
		chirps, err = cfg.dbQueries.GetAllChirps(r.Context())
		if err != nil {
			respondWithError(w, 400, fmt.Sprint(err))
			return
		}
	} else {
		user_id, err := uuid.Parse(author_id)
		if err != nil {
			respondWithError(w, 400, err.Error())
		}
		chirps, err = cfg.dbQueries.GetChirpsOfUser(r.Context(), user_id)
		if err != nil {
			respondWithError(w, 400, err.Error())
		}
	}
	var dat []ChirpJson
	for _, chirp := range chirps {
		dat = append(dat, ChirpJson{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID})
	}
	if sort_param == "desc" {
		slices.Reverse(dat)
	}
	respondWithJson(w, 200, dat)
}

func (cfg *apiConfig) handleChirpById(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	fmt.Print(id)
	if err != nil {
		respondWithError(w, 400, fmt.Sprint(err))
		return
	}
	chirp, err := cfg.dbQueries.GetChirpById(r.Context(), id)
	cj := ChirpJson{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}
	respondWithJson(w, 200, cj)
}

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	jwtoken := r.Header.Get("Authorization")
	if jwtoken == "" {
		respondWithError(w, 403, "no jwt token")
		return
	}
	token := jwtoken[7:]
	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, fmt.Sprint(err))
		return
	}
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 400, fmt.Sprint(err))
		return
	}
	chirp, err := cfg.dbQueries.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 404, fmt.Sprint(err))
		return
	}
	if chirp.UserID != userId {
		respondWithError(w, 403, "not yours chirp")
		return
	}
	err = cfg.dbQueries.DeleteChirp(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 500, "failed to delete chirp")
		return
	}
	w.WriteHeader(204)
}
