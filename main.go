package main

import (
    "bufio"
//    "fmt"
    "net/http"
    "os/exec"
    "regexp"
    "strings"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "log"
)

var (
    activeUsers = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "active_users",
            Help: "Number of active users",
        },
        []string{"user", "ip_address", "logon_time"},
    )
)

func init() {
    prometheus.MustRegister(activeUsers)
}

func recordMetrics() {
    go func() {
        for {
            collectActiveUsers()
            time.Sleep(10 * time.Second)
        }
    }()
}

func collectActiveUsers() {
    // Run the `w` command
    cmd := exec.Command("w", "-h")
    stdout, err := cmd.Output()
    if err != nil {
        log.Printf("Error running w command: %v", err)
        return
    }

    scanner := bufio.NewScanner(strings.NewReader(string(stdout)))
    userSessionRe := regexp.MustCompile(`^(\S+)\s+\S+\s+(\S+)\s+.*?\s+(\d+:\d+)`)

    // Clear previous metrics
    activeUsers.Reset()

    for scanner.Scan() {
        line := scanner.Text()
        matches := userSessionRe.FindStringSubmatch(line)
        if matches != nil {
            user := matches[1]
            ipAddress := matches[2]
            logonTime := matches[3]
            activeUsers.WithLabelValues(user, ipAddress, logonTime).Set(1)
        }
    }
}

func main() {
    recordMetrics()

    http.Handle("/metrics", promhttp.Handler())
    log.Println("Beginning to serve on port :8987")
    log.Fatal(http.ListenAndServe(":8987", nil))
}
