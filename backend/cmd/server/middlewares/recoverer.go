package middlewares

import (
	"fmt"

	"github.com/ogen-go/ogen/middleware"
)

func Recoverer(r middleware.Request, next middleware.Next) (resp middleware.Response, err error) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}

		resp = middleware.Response{}
		err = fmt.Errorf("recovered from panic: %v", rec)
	}()

	return next(r)
}
