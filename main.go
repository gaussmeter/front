package main

//Todo: implement SSL/https

import (
  "bytes"
  "fmt"
  "github.com/gorilla/handlers"
  "github.com/gorilla/mux"
  "log"
  "net/http"
  "net/http/httputil"
  "os"
)

type token_auth struct {
  Email string    `json:"email"`
  Password string `json:"password"`
}

func putValue(key string, value string) {
  _, err := http.Post("http://config:8443/badger/"+key, "text", bytes.NewBufferString(value))
  if err != nil {
    fmt.Printf("%s \n",err)
  } else {
    fmt.Printf("put key: %s value: %s", key, value)
  }

}

func putDefaults()  {
     putValue("default:tHome","37.4919392,-121.9469367")
     putValue("default:tHomeRadiusFt","100")
     putValue("default:tWork","37.4919392,-121.9469367")
     putValue("default:tWorkRadiusFt","100")
     putValue("default:tChargeRangeFull", "270")
     putValue("default:tChargeRangeMedium","100")
     putValue("default:tChargeRangeLow","30")
     putValue("default:eIHIP","fill")
     putValue("default:eIHNP","fill")
     putValue("default:eIHNPBCRM","fill")
     putValue("default:eNH","rainbow")
     putValue("default:cIHIP","0,0,0,255")
     putValue("default:cIHNP","0,0,0,255")
     putValue("defautl:cIHNPBCRM","0,0,0,255")
     putValue("default:jcNH","0,0,0,255")
     putValue("default:tGetStateInterval","14400")
     putValue("default:tSoftStateInterval","600")
     putValue("default:tGetStateIntervalDriving","30")
     putValue("default:tGetStateIntervalCharging","60")
     putValue("default:tVehicle","0")
}

func main() {
  putDefaults()
  proxyBadger := &httputil.ReverseProxy{Director: func(req *http.Request) {
    //Todo: allow override via environment variables
    originHost := "config:8443"
    req.Header.Add("X-Forwarded-Host", req.Host)
    req.Header.Add("X-Origin-Host", originHost)
    req.Host = originHost
    req.URL.Scheme = "http"
    req.URL.Host = originHost
  }}
  proxyToken := &httputil.ReverseProxy{Director: func(req *http.Request) {
    //Todo: allow override via environment variables
    originHost := "token:9001"
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
  rtr.HandleFunc("/secret/{key}", func(w http.ResponseWriter, r *http.Request) {
    key := mux.Vars(r)["key"]
    //Todo: allow override via environment variables
    secret, err := http.Get("http://config:8443/secret/"+key)
    if err == nil && secret.StatusCode == 200 {
      w.WriteHeader(http.StatusOK)
      fmt.Fprintf(w,"...hidden...")
    } else {
      w.WriteHeader(http.StatusNotFound)
    }
  }).Methods("GET")
  rtr.HandleFunc("/tToken", func(w http.ResponseWriter, r *http.Request) {
    proxyToken.ServeHTTP(w, r)
  } ).Methods("POST")
  rtr.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
  loggedRouter := handlers.LoggingHandler(os.Stdout, rtr)

  log.Fatal(http.ListenAndServe(":9001", loggedRouter))
}