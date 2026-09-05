package controller

import (
	"net/http"

	"github.com/imsab23/platform-be/pkg/http/router"
)

// wrapNetHTTPMiddleware bridges a standard net/http middleware into a platform
// router.Middleware. Context mutations made by mw (e.g. identity injection) are
// propagated into downstream platform handlers via the updated *http.Request.
func wrapNetHTTPMiddleware(mw func(http.Handler) http.Handler) router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Ctx) error {
			var (
				called     bool
				handlerErr error
			)
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				// r carries the updated context (e.g. identity stored by auth middleware).
				newC := router.NewCtx(r.Context(), w, r, nil)
				handlerErr = next(newC)
			})
			mw(inner).ServeHTTP(c.ResponseWriter(), c.Request())
			if !called {
				// Middleware rejected the request and already wrote its response.
				return nil
			}
			return handlerErr
		}
	}
}
