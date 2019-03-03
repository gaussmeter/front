package main

//Todo: implement SSL/https

import (
  "github.com/gorilla/handlers"
  "github.com/gorilla/mux"
  "log"
  "net/http"
  "net/http/httputil"
  "os"
  "strings"
)

func main() {
  proxyBadger := &httputil.ReverseProxy{Director: func(req *http.Request) {
    //Todo: allow override via environment variables
    originHost := "config:8443"
    //Todo: change how badger handles secerts, remove this.
    if strings.Split(req.URL.Path, "/")[1] == "secret" {
      req.URL.Path = "/badger/" + strings.Split(req.URL.Path,"/")[2]

    }
    req.Header.Add("X-Forwarded-Host", req.Host)
    req.Header.Add("X-Origin-Host", originHost)
    req.Host = originHost
    req.URL.Scheme = "http"
    req.URL.Host = originHost
  }}

  rtr := mux.NewRouter()
  rtr.HandleFunc("/badger/{*}", func(w http.ResponseWriter, r *http.Request) {
    proxyBadger.ServeHTTP(w, r)
  }).Methods("GET","PUT","POST")
  rtr.HandleFunc("/streamr/{*}", func(w http.ResponseWriter, r *http.Request) {
    proxyBadger.ServeHTTP(w, r)
  }).Methods("GET")
  rtr.HandleFunc("/secret/{*}", func(w http.ResponseWriter, r *http.Request) {
    proxyBadger.ServeHTTP(w, r)
  }).Methods("PUT", "POST")
  rtr.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
  loggedRouter := handlers.LoggingHandler(os.Stdout, rtr)

  log.Fatal(http.ListenAndServe(":9001", loggedRouter))
}