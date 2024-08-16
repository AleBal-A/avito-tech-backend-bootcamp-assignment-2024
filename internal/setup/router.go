package setup

import (
	"avito/internal/custommiddleware"
	"avito/internal/handlers/authHandler"
	"avito/internal/handlers/flatHandler"
	"avito/internal/handlers/houseHandler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
)

func SetupRouter(
	authH authHandler.AuthHandler,
	houseH houseHandler.HouseHandler,
	flatH flatHandler.FlatHandler,
	logger *slog.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Get("/dummyLogin", authH.DummyLogin)
	r.Post("/register", authH.Register)
	r.Post("/login", authH.Login)

	// Protected routes moderationsOnly
	r.Group(func(r chi.Router) {
		r.Use(custommiddleware.AuthMiddleware(authH, logger))
		r.Use(custommiddleware.RoleMiddleware([]string{"moderator"}, logger))

		r.Post("/house/create", houseH.Create)
		r.Post("/flat/update", flatH.Update)
	})
	// Protected routes authOnly
	r.Group(func(r chi.Router) {
		r.Use(custommiddleware.AuthMiddleware(authH, logger))

		r.Get("/house/{id}", houseH.GetFlatsByHouseID)
		r.Post("/flat/create", flatH.Create)
	})

	return r
}
