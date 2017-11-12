package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

var (
	addr         = ":8182"
	apiPrefix    = "/api"
	publicPrefix = "/"
	publicDir    = "public"
)

func main() {
	startWeb(rootRouter())
}

func startWeb(h http.Handler) {
	srv := &http.Server{
		Addr:           addr,
		Handler:        h,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Listening on: %q", addr)
	log.Fatal(srv.ListenAndServe())
}

func rootRouter() *mux.Router {
	common := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
	)

	handler := mux.NewRouter()
	handler.PathPrefix(apiPrefix).Handler(common.With(negroni.Wrap(apiRouter(apiPrefix))))
	handler.PathPrefix(publicPrefix).Handler(common.With(negroni.NewStatic(http.Dir(publicDir))))

	return handler
}

func apiRouter(prefix string) *mux.Router {
	api := mux.NewRouter().PathPrefix(prefix).Subrouter().StrictSlash(true)
	api.HandleFunc("/location/", locationHandler).Methods("GET")
	return api
}

func locationHandler(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	loc := vars.Get("location")
	if loc == "" {
		l, err := LocationFrom(getIP(r))
		if err != nil {
			toJSONErr(w, err, http.StatusBadRequest)
			return
		}
		loc = l.String()
	}
	stores, err := StoresNear(loc)
	if err != nil {
		toJSONErr(w, err, http.StatusBadRequest)
		return
	}
	toJSON(w, stores)
}

func toJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		toJSONErr(w, err, http.StatusInternalServerError)
		return
	}
}

func toJSONErr(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"err":%q}`, err.Error())
}

func getIP(r *http.Request) string {
	vars := r.URL.Query()
	ip := vars.Get("ip")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")]
		}
	}
	return ip
}
