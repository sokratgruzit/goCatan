package getuser

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	resp "github.com/sokratgruzit/goCatan/internal/lib/api/response"
	"github.com/sokratgruzit/goCatan/internal/models"
	"github.com/sokratgruzit/goCatan/internal/storage"
)

type UserGetter interface {
	User(email string) (*models.User, error)
}

type Response struct {
	resp.Response
	User models.User `json:"user"`
}

func New(log *slog.Logger, userGetter UserGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.get.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		email := strings.TrimSpace(r.URL.Query().Get("email"))

		if email == "" {
			render.JSON(w, r, resp.Error("email query param required"))
			return
		}

		user, err := userGetter.User(email)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				render.JSON(w, r, resp.Error("user not found"))
				return
			}
			log.Error("failed to get user", slog.Any("err", err))
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		responseOK(w, r, user)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, user *models.User) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		User:     *user,
	})
}
