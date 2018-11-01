package health

import (
    "fmt"
    "github.com/ccremer/clustercode-worker/util"
    "net/http"
    "os"
)

func StartServer() {
    log.Info("Starting health server")
    http.HandleFunc("/", handleRoot)
    http.HandleFunc("/health", handleHealth)
    go func() {
        err := http.ListenAndServe(":8080", nil)
        util.PanicOnError(err)
    }()
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "This worker seems to be responding, but better check /health")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
    fileName := ".healthcheck"
    file,err := os.OpenFile(fileName, os.O_WRONLY, 0666)
    if err != nil {
        log.Warnf("%s", err)
    }
    file.Close()
}
