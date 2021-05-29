package main

import (
        "fmt"
        "encoding/json"
        "crypto/tls"
        "net/http"
        "github.com/gorilla/mux"
)

func httpapi() {
        router := GenerateRouter()
        addr := fmt.Sprintf(":%d", Global.Port)
        certFile := Global.ServerOptions.CertFile
        keyFile := Global.ServerOptions.KeyFile

        if Global.ServerOptions.EnableTLS {
                srv := &http.Server{
                        Addr:         addr,
                        Handler:      router,
                        TLSConfig:    &Global.ServerOptions.TLSConfig,
                        TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
                }
                log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
        } else {
        srv := &http.Server{
                        Addr:         addr,
                        Handler:      router,
                }
        srv.ListenAndServe()
        log.Info("http api server have been started without ssl")
        }
}

func JustRoot(w http.ResponseWriter, r *http.Request) {
}

func Ping(w http.ResponseWriter, r *http.Request) {    

}

func AddMessage(w http.ResponseWriter, r *http.Request) {
        text := r.FormValue("text")
        Item := QueueItem{Priority: 2, Type: 1, Text: text}
        AddToQueue(Item)

        ReturnResponse(w, 200, map[string]interface{}{
                "sent": true,
        })
}

func ReturnResponse(w http.ResponseWriter, statusCode int, resp interface{}) {
        bytes, _ := json.Marshal(resp)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(statusCode)
        _, _ = fmt.Fprintf(w, string(bytes))
}

func GenerateRouter() *mux.Router {
        r := mux.NewRouter()
        // Endpoints called external shit
        r.HandleFunc("/", JustRoot).Methods(http.MethodGet)
        r.HandleFunc("/api/send", AddMessage).Methods(http.MethodPost)
        r.HandleFunc("/api/ping", Ping).Methods(http.MethodPost)
        return r
}

