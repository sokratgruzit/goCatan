package getusers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	resp "github.com/sokratgruzit/goCatan/internal/lib/api/response"
	"github.com/sokratgruzit/goCatan/internal/models"
)

type UserListGetter interface {
	Users() ([]*models.User, error)
}

type Response struct {
	resp.Response
	Users []models.User `json:"users"`
}

func New(log *slog.Logger, userGetter UserListGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.getall.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		usersPtr, err := userGetter.Users()
		if err != nil {
			log.Error("failed to get users", slog.Any("err", err))
			render.JSON(w, r, resp.Error("failed to get users"))
			return
		}

		// Convert []*User to []User for JSON marshaling
		users := make([]models.User, 0, len(usersPtr))
		for _, u := range usersPtr {
			users = append(users, *u)
		}

		responseOK(w, r, users)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, users []models.User) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Users:    users,
	})
}
