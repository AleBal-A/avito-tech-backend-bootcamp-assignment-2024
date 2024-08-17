package avito_test

import (
	"avito/internal/domain/models"
	"avito/internal/handlers/authHandler"
	"avito/internal/handlers/flatHandler"
	"avito/internal/handlers/houseHandler"
	"avito/internal/handlers/response"
	"avito/internal/repositories/mocks"
	"avito/internal/services/authService"
	"avito/internal/services/flatService"
	"avito/internal/services/houseService"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	logOff "log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"avito/internal/lib/logger"
	"avito/internal/setup"
)

func TestRegisterMod(t *testing.T) {
	log := logger.SetupLogger("debug")

	authRepoMock := mocks.NewAuthRepo(t)
	houseRepoMock := mocks.NewHouseRepo(t)
	flatRepoMock := mocks.NewFlatRepo(t)

	authRepoMock.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).
		Return("1", nil)

	authS := authService.NewService(authRepoMock, "jwt_secret", log)
	houseS := houseService.NewService(houseRepoMock, log)
	flatS := flatService.NewService(flatRepoMock, log)

	authH := authHandler.NewHandler(authS, log)
	houseH := houseHandler.NewHandler(houseS, log)
	flatH := flatHandler.NewHandler(flatS, log)

	router := setup.SetupRouter(authH, houseH, flatH, log)

	t.Run("Register moderator", func(t *testing.T) {
		body := `{
            "id_user": "moderator@example.com",
            "password": "pass",
            "user_type": "moderator"
        }`
		req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
	})
}

// Тесты для сценариев получения списка квартир
// /house/{id} "client"
func TestGetFlatsByHouseIDAsClient(t *testing.T) {
	log := logger.SetupLogger("debug")

	authRepoMock := mocks.NewAuthRepo(t)
	houseRepoMock := mocks.NewHouseRepo(t)
	flatRepoMock := mocks.NewFlatRepo(t)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("qwerty"), bcrypt.DefaultCost)

	authRepoMock.On("GetUserByEmail", mock.Anything, "client-uuid").
		Return(&models.User{
			ID:       "client-uuid",
			Password: string(hashedPassword),
			Role:     "client",
		}, nil)

	houseRepoMock.On("GetFlatsByHouseID", mock.Anything, 12345, "client").
		Return([]models.Flat{
			{
				ID:      123456,
				HouseID: 12345,
				Price:   10000,
				Rooms:   4,
				Status:  "approved",
			},
		}, nil)

	authS := authService.NewService(authRepoMock, "jwt_secret", log)
	houseS := houseService.NewService(houseRepoMock, log)
	flatS := flatService.NewService(flatRepoMock, log)

	authH := authHandler.NewHandler(authS, log)
	houseH := houseHandler.NewHandler(houseS, log)
	flatH := flatHandler.NewHandler(flatS, log)

	router := setup.SetupRouter(authH, houseH, flatH, log)

	t.Run("Login as client", func(t *testing.T) {
		body := `{
			"id": "client-uuid",
			"password": "qwerty"
		}`
		req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		userToken := extractTokenFromResponse(resp.Body.String())

		// Запрос квартир в доме как обычный пользователь
		req = httptest.NewRequest("GET", "/house/12345", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
		resp = httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var actualResponse map[string][]response.FlatResponse
		err := json.Unmarshal(resp.Body.Bytes(), &actualResponse)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		expectedResponse := []response.FlatResponse{
			{
				ID:      123456,
				HouseID: 12345,
				Price:   10000,
				Rooms:   4,
				Status:  "approved",
			},
		}

		assert.Equal(t, expectedResponse, actualResponse["flats"])

		houseRepoMock.AssertExpectations(t)
	})
}

// Тесты для сценариев получения списка квартир
// /house/{id} "moderator"
func TestGetFlatsByHouseIDAsModerator(t *testing.T) {
	log := logger.SetupLogger("debug")

	authRepoMock := mocks.NewAuthRepo(t)
	houseRepoMock := mocks.NewHouseRepo(t)
	flatRepoMock := mocks.NewFlatRepo(t)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("qwerty"), bcrypt.DefaultCost)

	authRepoMock.On("GetUserByEmail", mock.Anything, "moderator-uuid").
		Return(&models.User{
			ID:       "moderator-uuid",
			Password: string(hashedPassword),
			Role:     "moderator",
		}, nil)

	houseRepoMock.On("GetFlatsByHouseID", mock.Anything, 12345, "moderator").
		Return([]models.Flat{
			{
				ID:      123456,
				HouseID: 12345,
				Price:   10000,
				Rooms:   4,
				Status:  "approved",
			},
			{
				ID:      123457,
				HouseID: 12345,
				Price:   15000,
				Rooms:   5,
				Status:  "created",
			},
		}, nil)

	authS := authService.NewService(authRepoMock, "jwt_secret", log)
	houseS := houseService.NewService(houseRepoMock, log)
	flatS := flatService.NewService(flatRepoMock, log)

	authH := authHandler.NewHandler(authS, log)
	houseH := houseHandler.NewHandler(houseS, log)
	flatH := flatHandler.NewHandler(flatS, log)

	router := setup.SetupRouter(authH, houseH, flatH, log)

	t.Run("Login as moderator", func(t *testing.T) {
		body := `{
			"id": "moderator-uuid",
			"password": "qwerty"
		}`
		req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		moderatorToken := extractTokenFromResponse(resp.Body.String())

		req = httptest.NewRequest("GET", "/house/12345", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", moderatorToken))
		resp = httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		var actualResponse map[string][]response.FlatResponse
		err := json.Unmarshal(resp.Body.Bytes(), &actualResponse)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		expectedResponse := []response.FlatResponse{
			{
				ID:      123456,
				HouseID: 12345,
				Price:   10000,
				Rooms:   4,
				Status:  "approved",
			},
			{
				ID:      123457,
				HouseID: 12345,
				Price:   15000,
				Rooms:   5,
				Status:  "created",
			},
		}

		assert.Equal(t, expectedResponse, actualResponse["flats"])

		houseRepoMock.AssertExpectations(t)
	})
}

// тесты для сценариев публикации новой квартиры
func TestCreateHouseFlat(t *testing.T) {
	log := logger.SetupLogger("debug")

	authRepoMock := mocks.NewAuthRepo(t)
	houseRepoMock := mocks.NewHouseRepo(t)
	flatRepoMock := mocks.NewFlatRepo(t)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("qwerty"), bcrypt.DefaultCost)

	authRepoMock.On("GetUserByEmail", mock.Anything, "cae36e0f-69e5-4fa8-a179-a52d083c5549").
		Return(&models.User{
			ID:       "cae36e0f-69e5-4fa8-a179-a52d083c5549",
			Password: string(hashedPassword),
			Role:     "moderator",
		}, nil)

	houseRepoMock.On("CreateHouse", mock.Anything, mock.AnythingOfType("*models.House")).
		Return(nil)
	flatRepoMock.On("CreateFlat", mock.Anything, mock.AnythingOfType("*models.Flat")).
		Return(123456, nil)

	authS := authService.NewService(authRepoMock, "jwt_secret", log)
	houseS := houseService.NewService(houseRepoMock, log)
	flatS := flatService.NewService(flatRepoMock, log)

	authH := authHandler.NewHandler(authS, log)
	houseH := houseHandler.NewHandler(houseS, log)
	flatH := flatHandler.NewHandler(flatS, log)

	router := setup.SetupRouter(authH, houseH, flatH, log)

	var token string

	t.Run("Login as moderator", func(t *testing.T) {
		body := `{
            "id": "cae36e0f-69e5-4fa8-a179-a52d083c5549",
            "password": "qwerty"
        }`
		req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		token = extractTokenFromResponse(resp.Body.String())
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

		expectedResponse := map[string]interface{}{
			"id":         5,
			"address":    "Лесная улица, 7, Москва, 125196",
			"year":       2000,
			"developer":  "Мэрия города",
			"created_at": "2024-08-16T18:37:59.415689+03:00",
			"update_at":  "2024-08-16T18:37:59.415689+03:00",
		}

		var actualResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &actualResponse)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		assert.Equal(t, expectedResponse["address"], actualResponse["address"])
		assert.Equal(t, expectedResponse["year"], int(actualResponse["year"].(float64)))
		assert.Equal(t, expectedResponse["developer"], actualResponse["developer"])
	})

	t.Run("Create flat in house as moderator", func(t *testing.T) {
		body := `{
            "house_id": 5,
            "price": 10000,
            "rooms": 4
        }`
		req := httptest.NewRequest("POST", "/flat/create", strings.NewReader(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		expectedResponse := map[string]interface{}{
			"id":       123456,
			"house_id": 5,
			"price":    10000,
			"rooms":    4,
			"status":   "created",
		}

		var actualResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &actualResponse)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		assert.Equal(t, expectedResponse["house_id"], int(actualResponse["house_id"].(float64)))
		assert.Equal(t, expectedResponse["price"], int(actualResponse["price"].(float64)))
		assert.Equal(t, expectedResponse["rooms"], int(actualResponse["rooms"].(float64)))
		assert.Equal(t, expectedResponse["status"], actualResponse["status"])
	})
}

func TestUpdateFlatStatus(t *testing.T) {
	log := logger.SetupLogger("debug")

	authRepoMock := mocks.NewAuthRepo(t)
	houseRepoMock := mocks.NewHouseRepo(t)
	flatRepoMock := mocks.NewFlatRepo(t)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("qwerty"), bcrypt.DefaultCost)

	authRepoMock.On("GetUserByEmail", mock.Anything, "cae36e0f-69e5-4fa8-a179-a52d083c5549").
		Return(&models.User{
			ID:       "cae36e0f-69e5-4fa8-a179-a52d083c5549",
			Password: string(hashedPassword),
			Role:     "moderator",
		}, nil)

	flatRepoMock.On("UpdateFlatStatus", mock.Anything, 123456, "approved", mock.Anything).
		Return(&models.Flat{
			ID:      123456,
			HouseID: 12345,
			Price:   10000,
			Rooms:   4,
			Status:  "approved",
		}, nil)

	flatRepoMock.On("GetFlatByID", mock.Anything, 123456).
		Return(&models.Flat{
			ID:      123456,
			HouseID: 12345,
			Price:   10000,
			Rooms:   4,
			Status:  "created",
		}, nil)

	authS := authService.NewService(authRepoMock, "jwt_secret", log)
	houseS := houseService.NewService(houseRepoMock, log)
	flatS := flatService.NewService(flatRepoMock, log)

	authH := authHandler.NewHandler(authS, log)
	houseH := houseHandler.NewHandler(houseS, log)
	flatH := flatHandler.NewHandler(flatS, log)

	router := setup.SetupRouter(authH, houseH, flatH, log)

	var token string

	t.Run("Login as moderator", func(t *testing.T) {
		log.Debug("Starting login test")
		body := `{
            "id": "cae36e0f-69e5-4fa8-a179-a52d083c5549",
            "password": "qwerty"
        }`
		req := httptest.NewRequest("POST", "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		token = extractTokenFromResponse(resp.Body.String())

		authRepoMock.AssertExpectations(t)
	})

	t.Run("Update flat status as moderator", func(t *testing.T) {
		log.Debug("Starting update flat status test")
		body := `{
            "id": 123456,
            "status": "approved"
        }`
		req := httptest.NewRequest("POST", "/flat/update", strings.NewReader(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		log.Debug("Before calling flat update handler")
		router.ServeHTTP(resp, req)
		log.Debug("After calling flat update handler")

		assert.Equal(t, http.StatusOK, resp.Code)

		expectedResponse := map[string]interface{}{
			"id":       123456,
			"house_id": 12345,
			"price":    10000,
			"rooms":    4,
			"status":   "approved",
		}

		var actualResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &actualResponse)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		assert.Equal(t, expectedResponse["id"], int(actualResponse["id"].(float64)))
		assert.Equal(t, expectedResponse["house_id"], int(actualResponse["house_id"].(float64)))
		assert.Equal(t, expectedResponse["price"], int(actualResponse["price"].(float64)))
		assert.Equal(t, expectedResponse["rooms"], int(actualResponse["rooms"].(float64)))
		assert.Equal(t, expectedResponse["status"], actualResponse["status"])

		log.Debug("Flat update successful, response validated")

		flatRepoMock.AssertExpectations(t)
	})
}

func extractTokenFromResponse(response string) string {
	var result map[string]string
	err := json.Unmarshal([]byte(response), &result)
	if err != nil {
		logOff.Println("extractTokenFromResponse: err", err)
		return ""
	}
	return result["token"]
}
