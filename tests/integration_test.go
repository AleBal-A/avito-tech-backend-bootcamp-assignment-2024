package avito_test

import (
	"avito/internal/domain/models"
	"avito/internal/handlers/authHandler"
	"avito/internal/handlers/flatHandler"
	"avito/internal/handlers/houseHandler"
	"avito/internal/repositories/mocks"
	"avito/internal/services/authService"
	"avito/internal/services/flatService"
	"avito/internal/services/houseService"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"avito/internal/lib/logger"
	"avito/internal/setup"
)

var token string

func TestRegisterMod(t *testing.T) {

	log := logger.SetupLogger("debug")

	authRepoMock := mocks.NewAuthRepo(t)
	houseRepoMock := mocks.NewHouseRepo(t)
	flatRepoMock := mocks.NewFlatRepo(t)

	authRepoMock.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).
		Return("1", nil)

	authService := authService.NewService(authRepoMock, "jwt_secret", log)
	houseService := houseService.NewService(houseRepoMock, log)
	flatService := flatService.NewService(flatRepoMock, log)

	authH := authHandler.NewHandler(authService, log)
	houseH := houseHandler.NewHandler(houseService, log)
	flatH := flatHandler.NewHandler(flatService, log)

	router := setup.SetupRouter(authH, houseH, flatH, log)

	t.Run("Register moderator", func(t *testing.T) {
		body := `{
            "email": "2moderator@example.com",
            "password": "pass",
            "role": "moderator"
        }`
		req := httptest.NewRequest("POST", "/register", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
	})
}

func TestCreateHouse(t *testing.T) {

	log := logger.SetupLogger("debug")

	authRepoMock := mocks.NewAuthRepo(t)
	houseRepoMock := mocks.NewHouseRepo(t)
	flatRepoMock := mocks.NewFlatRepo(t)

	authRepoMock.On("GetUserByEmail", mock.Anything, "moderator@example.com").
		Return(&models.User{
			ID:       "1",
			Email:    "moderator@example.com",
			Password: "$2a$10$47qXfRNpeKHpymG5zOfy9uXTRbbHL5hwlIXq9uKz1Kg4vdRVpnMUa",
			Role:     "moderator",
		}, nil)

	houseRepoMock.On("CreateHouse", mock.Anything, mock.AnythingOfType("*models.House")).
		Return(nil)

	flatRepoMock.On("CreateFlat", mock.Anything, mock.AnythingOfType("*models.Flat")).
		Return(5, nil)

	authService := authService.NewService(authRepoMock, "jwt_secret", log)
	houseService := houseService.NewService(houseRepoMock, log)
	flatService := flatService.NewService(flatRepoMock, log)

	authH := authHandler.NewHandler(authService, log)
	houseH := houseHandler.NewHandler(houseService, log)
	flatH := flatHandler.NewHandler(flatService, log)

	router := setup.SetupRouter(authH, houseH, flatH, log)

	t.Run("Login as moderator", func(t *testing.T) {
		body := `{
            "email": "moderator@example.com",
            "password": "pass"
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
			"ID":            5,
			"Address":       "Лесная улица, 7, Москва, 125196",
			"YearBuilt":     2000,
			"Builder":       "Мэрия города",
			"CreatedAt":     "2024-08-16T18:37:59.415689+03:00",
			"LastFlatAdded": nil,
		}

		var actualResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &actualResponse)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		assert.Equal(t, expectedResponse["Address"], actualResponse["Address"])

		if yearBuilt, ok := actualResponse["YearBuilt"].(float64); ok {
			assert.Equal(t, expectedResponse["YearBuilt"], int(yearBuilt))
		} else {
			t.Errorf("YearBuilt is not of type float64: %v", actualResponse["YearBuilt"])
		}

		assert.Equal(t, expectedResponse["Builder"], actualResponse["Builder"])

		assert.Contains(t, actualResponse["CreatedAt"], "0001-01-01")
		assert.Nil(t, actualResponse["LastFlatAdded"])
	})

	// Создание квартиры
	t.Run("Create flat as client", func(t *testing.T) {
		body := `{"house_id":11,"flat_number":10,"price":200000,"rooms":2}`
		req := httptest.NewRequest("POST", "/flat/create", strings.NewReader(body))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)

		expectedResponse := map[string]interface{}{
			"ID":          5,
			"HouseID":     11,
			"FlatNumber":  10,
			"Price":       200000,
			"Rooms":       2,
			"Status":      "created",
			"ModeratorID": nil,
		}

		var actualResponse map[string]interface{}
		err := json.Unmarshal(resp.Body.Bytes(), &actualResponse)
		if err != nil {
			t.Fatal("Failed to unmarshal response:", err)
		}

		assert.Equal(t, expectedResponse["HouseID"], int(actualResponse["HouseID"].(float64)))
		assert.Equal(t, expectedResponse["FlatNumber"], int(actualResponse["FlatNumber"].(float64)))
		assert.Equal(t, expectedResponse["Price"], int(actualResponse["Price"].(float64)))
		assert.Equal(t, expectedResponse["Rooms"], int(actualResponse["Rooms"].(float64)))
		assert.Equal(t, expectedResponse["Status"], actualResponse["Status"])
		assert.Nil(t, actualResponse["ModeratorID"])
	})
}

func extractTokenFromResponse(response string) string {
	var result map[string]string
	json.Unmarshal([]byte(response), &result)
	return result["token"]
}
