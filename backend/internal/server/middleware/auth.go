package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
	"github.com/GaryHY/leviosa/pkg/contextutil"
	"github.com/GaryHY/leviosa/pkg/serverutil"
)

type ContextKey int

const UserIDKey = ContextKey(23)

// Function middleware to authenticate and authorize users.
func Auth(sessionGetter sessionGetterFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// make exception for certain path where you just call next.ServeHTTP(w,r)
			noAuthEndpoints := []string{
				"hello",
				"upload-image",

				"user/register",
				"user/validate-otp",
				"user/approve-user",

				"oauth/google/user",
			}
			var url string
			if r.URL.Path == "/favicon.ico" {
				fmt.Println("here in the favicon thing brother.")
				next.ServeHTTP(w, r)
				return
			} else {
				url = strings.Join(strings.Split(r.URL.Path, "/")[3:], "/")
				for _, endpoint := range noAuthEndpoints {
					if url == endpoint {
						next.ServeHTTP(w, r)
						return
					}
				}
			}
			ctx := r.Context()

			logger, ok := ctx.Value(contextutil.LoggerKey).(*slog.Logger)
			if !ok {
				http.Error(w, "logger not found in context", http.StatusInternalServerError)
				return
			}

			expectedRole := getExpectedRoleFromRequest(r)
			// get sessionID from request
			sessionID, err := getSessionIDFromRequest(r)
			if err != nil {
				logger.ErrorContext(ctx, "failed to get sessionID from the request", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// get session object from session repo
			session, err := sessionGetter(ctx, sessionID)
			if err != nil {
				logger.ErrorContext(ctx, "get session from database", "error", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// validate session
			if err := session.Valid(ctx, expectedRole); err != nil {
				logger.ErrorContext(ctx, "failed to validate session", "error", err)
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// next(w, r.WithContext(ctx))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TODO: find a way to get all the routes from an instance of server.
var knownBasicEndpoints = []string{
	"me",
	serverutil.SIGNINENDPOINT,
	serverutil.SIGNUPENDPOINT,
	serverutil.SIGNOUTENDPOINT,
	"auth",
	"logout",
	"vote",
	"checkout",
	"register",
	"event",
}

func getExpectedRoleFromRequest(r *http.Request) models.Role {
	segment := strings.Split(r.URL.Path, "/")[3]
	for _, endpoint := range knownBasicEndpoints {
		if segment == endpoint {
			return models.ConvertToRole("basic")
		}
	}
	return models.ConvertToRole(segment)
}
