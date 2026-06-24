package handlers

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "strings"

    "taskapi/db"
    "taskapi/models"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, map[string]string{"error": msg})
}

func parseID(r *http.Request) (int, error) {
    parts := strings.Split(r.URL.Path, "/")
    return strconv.Atoi(parts[len(parts)-1])
}

// ── GET /tasks ────────────────────────────────────────────────────────────────

func GetAllTasks(w http.ResponseWriter, r *http.Request) {
    rows, err := db.DB.Query(context.Background(),
        "SELECT id, title, done, created_at FROM tasks ORDER BY id")
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to fetch tasks")
        return
    }
    defer rows.Close()

    var tasks []models.Task
    for rows.Next() {
        var t models.Task
        if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
            writeError(w, http.StatusInternalServerError, "Failed to read task row")
            return
        }
        tasks = append(tasks, t)
    }

    // Return empty array [] instead of null when no tasks exist
    if tasks == nil {
        tasks = []models.Task{}
    }

    writeJSON(w, http.StatusOK, tasks)
}

// ── GET /tasks/{id} ───────────────────────────────────────────────────────────

func GetTask(w http.ResponseWriter, r *http.Request) {
    id, err := parseID(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, "Invalid task ID")
        return
    }

    var t models.Task
    err = db.DB.QueryRow(context.Background(),
        "SELECT id, title, done, created_at FROM tasks WHERE id = $1", id).
        Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)

    if err != nil {
        writeError(w, http.StatusNotFound, fmt.Sprintf("Task %d not found", id))
        return
    }

    writeJSON(w, http.StatusOK, t)
}

// ── POST /tasks ───────────────────────────────────────────────────────────────

func CreateTask(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Title string `json:"title"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid JSON body")
        return
    }
    if strings.TrimSpace(input.Title) == "" {
        writeError(w, http.StatusBadRequest, "Title cannot be empty")
        return
    }

    var t models.Task
    err := db.DB.QueryRow(context.Background(),
        `INSERT INTO tasks (title) VALUES ($1)
         RETURNING id, title, done, created_at`,
        input.Title).
        Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)

    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to create task")
        return
    }

    writeJSON(w, http.StatusCreated, t)
}

// ── PUT /tasks/{id} ───────────────────────────────────────────────────────────

func UpdateTask(w http.ResponseWriter, r *http.Request) {
    id, err := parseID(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, "Invalid task ID")
        return
    }

    var input struct {
        Title *string `json:"title"` // pointer = optional field
        Done  *bool   `json:"done"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        writeError(w, http.StatusBadRequest, "Invalid JSON body")
        return
    }

    // Fetch current task first
    var t models.Task
    err = db.DB.QueryRow(context.Background(),
        "SELECT id, title, done, created_at FROM tasks WHERE id = $1", id).
        Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
    if err != nil {
        writeError(w, http.StatusNotFound, fmt.Sprintf("Task %d not found", id))
        return
    }

    // Only update fields that were sent
    if input.Title != nil {
        t.Title = *input.Title
    }
    if input.Done != nil {
        t.Done = *input.Done
    }

    _, err = db.DB.Exec(context.Background(),
        "UPDATE tasks SET title = $1, done = $2 WHERE id = $3",
        t.Title, t.Done, t.ID)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to update task")
        return
    }

    writeJSON(w, http.StatusOK, t)
}

// ── DELETE /tasks/{id} ────────────────────────────────────────────────────────

func DeleteTask(w http.ResponseWriter, r *http.Request) {
    id, err := parseID(r)
    if err != nil {
        writeError(w, http.StatusBadRequest, "Invalid task ID")
        return
    }

    result, err := db.DB.Exec(context.Background(),
        "DELETE FROM tasks WHERE id = $1", id)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "Failed to delete task")
        return
    }

    // RowsAffected tells us if the ID actually existed
    if result.RowsAffected() == 0 {
        writeError(w, http.StatusNotFound, fmt.Sprintf("Task %d not found", id))
        return
    }

    writeJSON(w, http.StatusOK, map[string]string{
        "message": fmt.Sprintf("Task %d deleted successfully", id),
    })
}