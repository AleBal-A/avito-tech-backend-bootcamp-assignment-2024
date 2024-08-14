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

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(custommiddleware.AuthMiddleware(authH, logger))

		r.Post("/house/create", custommiddleware.RoleMiddleware([]string{"moderator"}, logger)(houseH.Create))
		//r.Get("/house/{id}", )

		r.Post("/flat/create", custommiddleware.RoleMiddleware([]string{"client", "moderator"}, logger)(flatH.Create))
		r.Post("/flat/{id}/update", custommiddleware.RoleMiddleware([]string{"moderator"}, logger)(flatH.Update))
	})

	return r
}
