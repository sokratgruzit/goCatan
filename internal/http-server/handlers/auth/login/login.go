package login

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	resp "github.com/sokratgruzit/goCatan/internal/lib/api/response"
	"github.com/sokratgruzit/goCatan/internal/models"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
	User models.User `json:"user"`
}

type UserLoginer interface {
	Login(email string, password string) (*models.User, error)
}

func New(log *slog.Logger, userLoginer UserLoginer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.login.New"

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

		user, err := userLoginer.Login(req.Email, req.Password)

		if err != nil {
			log.Error("failed to login user", slog.Any("err", err))
			render.JSON(w, r, resp.Error("invalid email or password"))
			return
		}

		log.Info("user logged in", slog.Any("user", user))

		responseOK(w, r, user)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, user *models.User) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		User:     *user,
	})
}
