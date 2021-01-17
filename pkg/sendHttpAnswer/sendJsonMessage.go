package sendHttpAnswer

import (
	"encoding/json"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"net/http"
)

func SendJson(answerStruct interface{}, ctx *routing.Context, responseStatus int) {
	jsonAnswer, err := json.Marshal(answerStruct)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(responseStatus)
	ctx.SetBody(jsonAnswer)
}

func SendJsonWithRequestContext(answerStruct interface{}, ctx *fasthttp.RequestCtx, responseStatus int) {
	jsonAnswer, err := json.Marshal(answerStruct)
	if err != nil {
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(responseStatus)
	ctx.SetBody(jsonAnswer)
}
