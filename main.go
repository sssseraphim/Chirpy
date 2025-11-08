package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	var cfg apiConfig
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.showHits)
	mux.HandleFunc("POST /admin/reset", cfg.resetHits)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	server := &http.Server{Handler: mux, Addr: ":8080"}
	server.ListenAndServe()
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "something went wrong")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	type ans struct {
		Body string `json:"cleaned_body"`
	}
	data, err := json.Marshal(ans{Body: cleanBody(params.Body)})
	if err != nil {
		fmt.Printf("Something went wrong")
	}
	resopondWithJson(w, 200, data)
	w.Write(data)
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

func resopondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Fatal("very bad")
	}
	w.Write(dat)
}
