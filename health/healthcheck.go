package health

import (
    "fmt"
    "github.com/ccremer/clustercode-worker/messaging"
    "github.com/ccremer/clustercode-worker/util"
    "github.com/micro/go-config"
    "net/http"
    "os"
)

type HealthCheckDto struct {
    InputDir  string `json:"input_dir"`
    OutputDir string `json:"output_dir"`
    Messaging string `json:"messaging"`
}

func StartServer() {
    port := config.Get("health", "port").String("8080")
    log.Infof("Starting health server on port %s", port)
    http.HandleFunc("/", handleHealth)
    go func() {
        err := http.ListenAndServe(":"+port, nil)
        util.PanicOnError(err)
    }()
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
    dto := HealthCheckDto{}
    dto.OutputDir = checkOutputDir(config.Get("output", "dir").String("/output") + "/.clustercode-health")
    dto.InputDir = checkInputDir(config.Get("input", "dir").String("/input") + "/0")
    dto.Messaging = "unknown"
    json, _ := messaging.ToJson(dto)
    fmt.Fprint(w, json)
    log.Debugf("response: %s", json)
}

func checkOutputDir(path string) string {
    file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
    defer file.Close()
    if err != nil {
        log.Warnf("%s", err)
        return fmt.Sprint(err)
    } else {
        os.Remove(path)
        return "ok"
    }
}

func checkInputDir(path string) string {
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return fmt.Sprint(err)
    } else {
        return "ok"
    }
}
