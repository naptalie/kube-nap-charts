package mid

import (
	"context"
	"errors"
	"net/http"

	"health-api/app/sdk/errs"
	"health-api/foundation/logger"
	"health-api/foundation/web"
)

// Errors handles errors from handlers.
func Errors(log *logger.Logger) web.Middleware {
	m := func(handler web.HandlerFunc) web.HandlerFunc {
		h := func(ctx context.Context, r *http.Request) web.Encoder {
			resp := handler(ctx, r)

			if !checkIsError(resp) {
				return resp
			}

			var appErr *errs.Error
			if !errors.As(resp.(error), &appErr) {
				appErr = errs.Newf(errs.Internal, "internal server error")
			}

			log.Error(ctx, "error handling request",
				"code", appErr.Code.String(),
				"message", appErr.Message,
				"source", appErr.FileName,
				"func", appErr.FuncName,
			)

			// Don't expose internal-only errors to clients
			if appErr.Code == errs.InternalOnlyLog {
				appErr = errs.Newf(errs.Internal, "internal server error")
			}

			return appErr
		}
		return h
	}
	return m
}
