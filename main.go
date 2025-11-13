package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sssseraphim/Chirpy/internal/database"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secret         string
	polkaKey       string
}

func main() {
	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	var cfg apiConfig
	cfg.platform = os.Getenv("PLATFORM")
	cfg.secret = os.Getenv("SECRET")
	cfg.polkaKey = os.Getenv("POLKA_KEY")

	cfg.dbQueries = database.New(db)
	const filepathRoot = "."
	const port = ":8080"
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.showHits)
	mux.HandleFunc("POST /admin/reset", cfg.resetUsers)
	mux.HandleFunc("POST /api/users", cfg.handleUsers)
	mux.HandleFunc("POST /api/chirps", cfg.handleChirps)
	mux.HandleFunc("GET /api/chirps", cfg.handleListChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handleChirpById)
	mux.HandleFunc("POST /api/login", cfg.handleLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handleRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handleRevoke)
	mux.HandleFunc("PUT /api/users", cfg.handleChangeEmailAndPassword)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handleDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.handleWebhooks)

	server := &http.Server{Handler: mux, Addr: port}
	err = server.ListenAndServe()
	fmt.Println(err)
}

func cleanBody(s string) string {
	bad := map[string]bool{"nigga": true, "fuck": true, "gay": true}
	words := strings.Split(s, " ")
	for i := range len(words) {
		if bad[strings.ToLower(words[i])] {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (a *apiConfig) middlewareMetricInc(next http.Handler) http.Handler {
	fmt.Printf("Hit")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) showHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, a.fileserverHits.Load())))
}

func (a *apiConfig) resetHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Reset!")
	a.fileserverHits.Store(0)
}

type errMessage struct {
	Error string `json:"error"`
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	e := errMessage{Error: msg}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	dat, err := json.Marshal(e)
	if err != nil {
		log.Fatal("bad")
	}
	w.Write(dat)
}

func respondWithJson(w http.ResponseWriter, code int, payload any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Fatal("very bad")
	}
	w.Write(dat)
}
