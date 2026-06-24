package main

import (
    "fmt"
    "log"
    "net/http"

    "taskapi/db"
    "taskapi/handlers"
    "taskapi/models"
)

func router(w http.ResponseWriter, r *http.Request) {
    // Route: /tasks or /tasks/{id}
    switch {
    case r.URL.Path == "/tasks":
        switch r.Method {
        case http.MethodGet:
            handlers.GetAllTasks(w, r)
        case http.MethodPost:
            handlers.CreateTask(w, r)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }

    case len(r.URL.Path) > 7 && r.URL.Path[:7] == "/tasks/":
        switch r.Method {
        case http.MethodGet:
            handlers.GetTask(w, r)
        case http.MethodPut:
            handlers.UpdateTask(w, r)
        case http.MethodDelete:
            handlers.DeleteTask(w, r)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }

    default:
        http.Error(w, "Not found", http.StatusNotFound)
    }
}

func main() {
    // 1. Connect to PostgreSQL
    db.Connect()

    // 2. Auto-create table if it doesn't exist
    if err := models.CreateTable(); err != nil {
        log.Fatalf("❌ Failed to create table: %v", err)
    }
    fmt.Println("✅ Table ready!")

    // 3. Register routes
    http.HandleFunc("/tasks", router)
    http.HandleFunc("/tasks/", router)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok"}`))
    })

    // 4. Start server
    fmt.Println("🚀 Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}