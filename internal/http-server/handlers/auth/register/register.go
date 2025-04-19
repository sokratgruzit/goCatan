package register

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	resp "github.com/sokratgruzit/goCatan/internal/lib/api/response"
	"github.com/sokratgruzit/goCatan/internal/storage"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Username string `json:"username" validate:"required"`
}

type Response struct {
	resp.Response
	UserID int64 `json:"user_id,omitempty"`
}

type UserRegisterer interface {
	RegisterUser(email string, password string, username string) (int64, error)
}

func New(log *slog.Logger, userRegisterer UserRegisterer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.register.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("failed to decode request body", slog.Any("err", err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", slog.Any("error", err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		userID, err := userRegisterer.RegisterUser(req.Email, req.Password, req.Username)

		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", slog.String("email", req.Email))
			render.JSON(w, r, resp.Error("user already exists"))
			return
		}

		if err != nil {
			log.Error("failed to register user", slog.Any("err", err))
			render.JSON(w, r, resp.Error("failed to register user"))
			return
		}

		log.Info("user registered", slog.Int64("user_id", userID))

		responseOK(w, r, userID)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, userID int64) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		UserID:   userID,
	})
}
