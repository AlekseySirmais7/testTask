package http

import (
	routing "github.com/qiangxue/fasthttp-routing"
	"net/http"
	"strconv"
	"testTask/internal/models"
	"testTask/internal/pkg/constants"
	"testTask/internal/pkg/product"
	"testTask/pkg/Random"
	"testTask/pkg/sendHttpAnswer"
)

type ProductHandler struct {
	ProductUseCase product.IUseCase
}

func (ph *ProductHandler) AddProducts(ctx *routing.Context) {

	sellerIdStr := string(ctx.QueryArgs().Peek("seller_id"))
	if sellerIdStr == "" {
		errStruct := models.Error{Message: `U need add GET parameter "seller_id"`}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
		return
	}
	sellerId, err := strconv.Atoi(sellerIdStr)
	if err != nil || sellerId < 0 {
		errStruct := models.Error{Message: `"seller_id" must be uint`}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
		return
	}

	xlsxUri := string(ctx.QueryArgs().Peek("xlsx_uri"))
	if xlsxUri == "" {
		errStruct := models.Error{Message: `U need add GET parameter "xlsx_uri"`}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
		return
	}

	statisticStruct, err := ph.ProductUseCase.AddProducts(sellerId, xlsxUri)
	if err != nil {
		errStruct := models.Error{Message: err.Error()}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
		return
	}
	sendHttpAnswer.SendJson(statisticStruct, ctx, http.StatusOK)
}

func (ph *ProductHandler) GetProducts(ctx *routing.Context) {

	sellerId := -1

	sellerIdStr := string(ctx.QueryArgs().Peek("seller_id"))
	if sellerIdStr != "" {
		sellerIdInt, err := strconv.Atoi(sellerIdStr)
		if err != nil || sellerIdInt < 0 {
			errStruct := models.Error{Message: `"seller_id" must be uint`}
			sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
			return
		}
		sellerId = sellerIdInt
	}

	offerId := -1

	offerIdStr := string(ctx.QueryArgs().Peek("offer_id"))
	if offerIdStr != "" {
		offerIdInt, err := strconv.Atoi(offerIdStr)
		if err != nil || offerIdInt < 0 {
			errStruct := models.Error{Message: `"offer_id" must be uint`}
			sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
			return
		}
		offerId = offerIdInt
	}

	name := string(ctx.QueryArgs().Peek("name"))

	products, err := ph.ProductUseCase.GetProducts(sellerId, offerId, name)
	if err != nil {
		errStruct := models.Error{Message: err.Error()}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
	}
	sendHttpAnswer.SendJson(products, ctx, http.StatusOK)
}

func (ph *ProductHandler) AddProductsAsync(ctx *routing.Context) {

	sellerIdStr := string(ctx.QueryArgs().Peek("seller_id"))
	if sellerIdStr == "" {
		errStruct := models.Error{Message: `U need add GET parameter "seller_id"`}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
		return
	}
	sellerId, err := strconv.Atoi(sellerIdStr)
	if err != nil || sellerId < 0 {
		errStruct := models.Error{Message: `"seller_id" must be uint`}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
		return
	}

	xlsxUri := string(ctx.QueryArgs().Peek("xlsx_uri"))
	if xlsxUri == "" {
		errStruct := models.Error{Message: `U need add GET parameter "xlsx_uri"`}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
		return
	}

	taskId, err := Random.GetRandomInt(constants.MaxTaskId)
	if err != nil {
		errStruct := models.Error{Message: `We have some Random err, repeat request pls`}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusInternalServerError)
		return
	}

	sendHttpAnswer.SendJson(models.TaskId{TaskId: taskId}, ctx, http.StatusOK)

	go ph.ProductUseCase.AddProductsAsync(sellerId, xlsxUri, taskId)
}

func (ph *ProductHandler) GetTaskStatus(ctx *routing.Context) {

	taskIdStr := string(ctx.QueryArgs().Peek("task_id"))
	if taskIdStr == "" {
		errStruct := models.Error{Message: `U need add GET parameter "task_id"`}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
		return
	}
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil || taskId < 0 {
		errStruct := models.Error{Message: `"task_id" must be uint`}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
		return
	}

	status, err := ph.ProductUseCase.GetTaskStatus(taskId)
	if err != nil {
		errStruct := models.Error{Message: `err get task status:` + err.Error()}
		sendHttpAnswer.SendJson(errStruct, ctx, http.StatusBadRequest)
		return
	}

	statusStruct := models.Status{TaskId: taskId, Status: status}
	sendHttpAnswer.SendJson(statusStruct, ctx, http.StatusOK)
}
