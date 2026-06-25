package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"taskapi/db"
	"taskapi/handlers"
	"taskapi/middleware"
	"taskapi/models"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	protected := r.Group("/tasks")
	protected.Use(middleware.AuthRequired())
	{
		protected.GET("", handlers.GetAllTasks)
		protected.POST("", handlers.CreateTask)
		protected.GET("/:id", handlers.GetTask)
		protected.PUT("/:id", handlers.UpdateTask)
		protected.DELETE("/:id", handlers.DeleteTask)
	}
	return r
}

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_URL", "postgres://postgres:123@localhost:5432/taskdb?sslmode=disable")

	db.Connect()
	models.CreateTable()
	models.CreateUsersTable()

	db.DB.Exec(context.Background(), "DELETE FROM tasks")
	db.DB.Exec(context.Background(), "DELETE FROM users WHERE username = 'testuser'")

	code := m.Run()

	db.DB.Exec(context.Background(), "DELETE FROM tasks")
	db.DB.Exec(context.Background(), "DELETE FROM users WHERE username = 'testuser'")

	os.Exit(code)
}

func makeRequest(r *gin.Engine, method, url, body, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	} else {
		reqBody = bytes.NewBufferString("")
	}

	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getToken(r *gin.Engine) string {
	makeRequest(r, "POST", "/register", `{"username":"testuser","password":"test1234"}`, "")
	w := makeRequest(r, "POST", "/login", `{"username":"testuser","password":"test1234"}`, "")

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp["token"]
}

func TestHealthCheck(t *testing.T) {
	r := setupRouter()
	w := makeRequest(r, "GET", "/health", "", "")

	assert.Equal(t, 200, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ok", resp["status"])

	fmt.Println("✅ TestHealthCheck passed")
}

func TestRegister(t *testing.T) {
	r := setupRouter()

	w := makeRequest(r, "POST", "/register", `{"username":"testuser","password":"test1234"}`, "")
	assert.Equal(t, 201, w.Code)

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Account created!", resp["message"])

	fmt.Println("✅ TestRegister passed")
}

func TestRegisterDuplicate(t *testing.T) {
	r := setupRouter()

	makeRequest(r, "POST", "/register", `{"username":"testuser","password":"test1234"}`, "")
	w := makeRequest(r, "POST", "/register", `{"username":"testuser","password":"test1234"}`, "")

	assert.Equal(t, 409, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Username already taken", resp["error"])

	fmt.Println("✅ TestRegisterDuplicate passed")
}

func TestLogin(t *testing.T) {
	r := setupRouter()
	makeRequest(r, "POST", "/register", `{"username":"testuser","password":"test1234"}`, "")

	w := makeRequest(r, "POST", "/login", `{"username":"testuser","password":"test1234"}`, "")
	assert.Equal(t, 200, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["token"])

	fmt.Println("✅ TestLogin passed")
}

func TestLoginWrongPassword(t *testing.T) {
	r := setupRouter()
	makeRequest(r, "POST", "/register", `{"username":"testuser","password":"test1234"}`, "")

	w := makeRequest(r, "POST", "/login", `{"username":"testuser","password":"wrongpassword"}`, "")
	assert.Equal(t, 401, w.Code)

	fmt.Println("✅ TestLoginWrongPassword passed")
}

func TestGetTasksWithoutToken(t *testing.T) {
	r := setupRouter()

	w := makeRequest(r, "GET", "/tasks", "", "")
	assert.Equal(t, 401, w.Code)

	fmt.Println("✅ TestGetTasksWithoutToken passed")
}

func TestGetTasksWithToken(t *testing.T) {
	r := setupRouter()
	token := getToken(r)

	w := makeRequest(r, "GET", "/tasks", "", token)
	assert.Equal(t, 200, w.Code)

	fmt.Println("✅ TestGetTasksWithToken passed")
}

func TestCreateTask(t *testing.T) {
	r := setupRouter()
	token := getToken(r)

	w := makeRequest(r, "POST", "/tasks", `{"title":"Test task"}`, token)
	assert.Equal(t, 201, w.Code)

	var task map[string]any
	json.Unmarshal(w.Body.Bytes(), &task)
	assert.Equal(t, "Test task", task["title"])
	assert.Equal(t, false, task["done"])

	fmt.Println("✅ TestCreateTask passed")
}

func TestCreateTaskEmptyTitle(t *testing.T) {
	r := setupRouter()
	token := getToken(r)

	w := makeRequest(r, "POST", "/tasks", `{"title":""}`, token)
	assert.Equal(t, 400, w.Code)

	fmt.Println("✅ TestCreateTaskEmptyTitle passed")
}

func TestUpdateTask(t *testing.T) {
	r := setupRouter()
	token := getToken(r)

	w := makeRequest(r, "POST", "/tasks", `{"title":"Update me"}`, token)
	var task map[string]any
	json.Unmarshal(w.Body.Bytes(), &task)
	id := int(task["id"].(float64))

	w = makeRequest(r, "PUT", fmt.Sprintf("/tasks/%d", id), `{"done":true}`, token)
	assert.Equal(t, 200, w.Code)

	var updated map[string]any
	json.Unmarshal(w.Body.Bytes(), &updated)
	assert.Equal(t, true, updated["done"])

	fmt.Println("✅ TestUpdateTask passed")
}

func TestDeleteTask(t *testing.T) {
	r := setupRouter()
	token := getToken(r)

	w := makeRequest(r, "POST", "/tasks", `{"title":"Delete me"}`, token)
	var task map[string]any
	json.Unmarshal(w.Body.Bytes(), &task)
	id := int(task["id"].(float64))

	w = makeRequest(r, "DELETE", fmt.Sprintf("/tasks/%d", id), "", token)
	assert.Equal(t, 200, w.Code)

	w = makeRequest(r, "GET", fmt.Sprintf("/tasks/%d", id), "", token)
	assert.Equal(t, 404, w.Code)

	fmt.Println("✅ TestDeleteTask passed")
}

func TestGetTaskNotFound(t *testing.T) {
	r := setupRouter()
	token := getToken(r)

	w := makeRequest(r, "GET", "/tasks/99999", "", token)
	assert.Equal(t, 404, w.Code)

	fmt.Println("✅ TestGetTaskNotFound passed")
}