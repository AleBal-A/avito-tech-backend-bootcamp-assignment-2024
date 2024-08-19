package avito_test

import (
	"avito/internal/handlers/authHandler"
	"avito/internal/handlers/flatHandler"
	"avito/internal/handlers/houseHandler"
	"avito/internal/lib/logger"
	"avito/internal/repositories/authRepo"
	"avito/internal/repositories/flatRepo"
	"avito/internal/repositories/houseRepo"
	"avito/internal/services/authService"
	"avito/internal/services/flatService"
	"avito/internal/services/houseService"
	"avito/internal/setup"
	"avito/internal/storage"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"avito/internal/config"
	_ "github.com/lib/pq"
)

var conn *sql.DB
var token string

func setupTestDB() *sql.DB {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5433",
			Name:     "test_db",
			User:     "test_user",
			Password: "qwertest",
		},
	}

	var err error
	conn, err = storage.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to the test database: %v", err)
	}

	log.Println("Successfully connected to the test database:", cfg.Database.Host)
	return conn
}

func teardownTestDB() {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			log.Fatalf("Failed to close the test database connection: %v", err)
		}
	}
}

func TestMain(m *testing.M) {
	conn = setupTestDB()
	code := m.Run()
	teardownTestDB()
	os.Exit(code)
}

func TestGetFlatsByIdWithRegistation(t *testing.T) {
	log := logger.SetupLogger("debug")

	authR := authRepo.NewRepository(conn, log)
	houseR := houseRepo.NewRepository(conn, log)
	flatR := flatRepo.NewRepository(conn, log)

	authS := authService.NewService(authR, "secret", log)
	houseS := houseService.NewService(houseR, log)
	flatS := flatService.NewService(flatR, log)

	authH := authHandler.NewHandler(authS, log)
	houseH := houseHandler.NewHandler(houseS, log)
	flatH := flatHandler.NewHandler(flatS, log)

	router := setup.SetupRouter(authH, houseH, flatH, log)

	var userID string
	var houseID int

	t.Run("Dummy login as client", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/dummyLogin?user_type=client", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		expectedSubstring := `"token":`
		assert.Contains(t, resp.Body.String(), expectedSubstring)
	})

	t.Run("Register moderator", func(t *testing.T) {
		body := `{
            "email": "moderator@example.com",
            "password": "pass",
            "user_type": "moderator"
        }`
		req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		userID = result["user_id"]
		assert.NotEmpty(t, userID, "Expected a valid user_id")
	})

	t.Run("Register client", func(t *testing.T) {
		body := `{
            "email": "client@example.com",
            "password": "pass",
            "user_type": "client"
        }`
		req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("Login as moderator", func(t *testing.T) {
		body := `{
            "id": "` + userID + `",
            "password": "pass"
        }`
		req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var result map[string]string
		err := json.Unmarshal(resp.Body.Bytes(), &result)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		token = result["token"]
		assert.NotEmpty(t, token, "Expected a valid token")
	})

	t.Run("Create house as moderator", func(t *testing.T) {
		body := `{
            "address": "Лесная улица, 7, Москва, 125196",
            "year": 2000,
            "developer": "Мэрия города"
        }`
		req := httptest.NewRequest("POST", "/house/create", strings.NewReader(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var houseResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &houseResponse)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		assert.Equal(t, "Лесная улица, 7, Москва, 125196", houseResponse["address"])
		assert.Equal(t, float64(2000), houseResponse["year"])
		assert.Equal(t, "Мэрия города", houseResponse["developer"])

		houseID = int(houseResponse["id"].(float64))
		assert.NotZero(t, houseID, "Expected a valid house ID")
	})

	t.Run("Create flat in house as moderator", func(t *testing.T) {
		body := fmt.Sprintf(`{
            "house_id": %d,
            "price": 10000,
            "rooms": 4
        }`, houseID)
		req := httptest.NewRequest("POST", "/flat/create", strings.NewReader(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var flatResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &flatResponse)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		assert.Equal(t, float64(houseID), flatResponse["house_id"])
		assert.Equal(t, float64(10000), flatResponse["price"])
		assert.Equal(t, float64(4), flatResponse["rooms"])
		assert.Equal(t, "created", flatResponse["status"])
	})

	t.Run("Get flats in house as moderator", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/house/%d", houseID), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var actualResponse map[string][]map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &actualResponse)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		flats, ok := actualResponse["flats"]
		assert.True(t, ok, "Expected 'flats' key in response")

		assert.NotEmpty(t, flats, "Expected at least one flat in response")
		assert.Equal(t, float64(houseID), flats[0]["house_id"])
		assert.Equal(t, float64(10000), flats[0]["price"])
		assert.Equal(t, float64(4), flats[0]["rooms"])
		assert.Equal(t, "created", flats[0]["status"])
	})
}
