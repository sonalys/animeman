package middlewares

import (
	"fmt"

	"github.com/ogen-go/ogen/middleware"
)

func Recoverer(r middleware.Request, next middleware.Next) (resp middleware.Response, err error) {
	defer func() {
		if r := recover(); r != nil {
			resp = middleware.Response{}
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	return next(r)
}
