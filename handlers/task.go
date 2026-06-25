package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"taskapi/db"
	"taskapi/models"
)

// GET /tasks
func GetAllTasks(c *gin.Context) {
	rows, err := db.DB.Query(context.Background(),
		"SELECT id, title, done, created_at FROM tasks ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read task"})
			return
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []models.Task{}
	}
	c.JSON(http.StatusOK, tasks)
}

// GET /tasks/:id
func GetTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var t models.Task
	err = db.DB.QueryRow(context.Background(),
		"SELECT id, title, done, created_at FROM tasks WHERE id = $1", id).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Task %d not found", id)})
		return
	}
	c.JSON(http.StatusOK, t)
}

// POST /tasks
func CreateTask(c *gin.Context) {
	var input struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}
	if strings.TrimSpace(input.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	var t models.Task
	err := db.DB.QueryRow(context.Background(),
		`INSERT INTO tasks (title) VALUES ($1)
		 RETURNING id, title, done, created_at`,
		input.Title).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}
	c.JSON(http.StatusCreated, t)
}

// PUT /tasks/:id
func UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var input struct {
		Title *string `json:"title"`
		Done  *bool   `json:"done"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	var t models.Task
	err = db.DB.QueryRow(context.Background(),
		"SELECT id, title, done, created_at FROM tasks WHERE id = $1", id).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Task %d not found", id)})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}
	c.JSON(http.StatusOK, t)
}

// DELETE /tasks/:id
func DeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	result, err := db.DB.Exec(context.Background(),
		"DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}
	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Task %d not found", id)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Task %d deleted", id)})
}