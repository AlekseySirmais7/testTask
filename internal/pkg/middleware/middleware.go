package middleware

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"math/rand"
	"net/http"
	"testTask/internal/models"
	"testTask/internal/pkg/constants"
	"testTask/pkg/sendHttpAnswer"
	"time"
)

type MiddlewareWithLogger struct {
	logger *zap.Logger
}

func NewMiddlewareWithLogger(logger *zap.Logger) MiddlewareWithLogger {
	return MiddlewareWithLogger{
		logger: logger,
	}
}

func (ml *MiddlewareWithLogger) PanicMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		defer func() {
			if err := recover(); err != nil {

				RequestId, ok := ctx.UserValue(constants.RequestIdKey).(string)
				if !ok {
					RequestId = "no id"
				}

				ml.logger.Error("PanicMiddleware",
					zap.String("RequestId:", fmt.Sprintf("%s", RequestId)),
					zap.String("Error:", fmt.Sprintf("%v", err)),
				)

				errStruct := models.Error{Message: `RecoverMiddleware =(, we will work soon. Report your request`}
				sendHttpAnswer.SendJsonWithRequestContext(errStruct, ctx, http.StatusInternalServerError)
				return
			}
		}()
		next(ctx)
	}
}

func (ml *MiddlewareWithLogger) LogMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		start := time.Now()
		requestId := fmt.Sprintf("%016x", rand.Int())[:10]

		ctx.SetUserValue(constants.RequestIdKey, requestId)

		defer func() {
			ml.logger.Info(string(ctx.Path()),
				zap.String("RequestId:", requestId),
				zap.String("Method:", string(ctx.Method())),
				zap.String("RemoteAddr:", ctx.RemoteAddr().String()),
				zap.Time("StartTime:", start),
				zap.Duration("DurationTime:", time.Since(start)),
			)
		}()
		next(ctx)
	}
}
