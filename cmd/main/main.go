package main

import (
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"log"
	"testTask/internal/pkg/logger"
	"testTask/internal/pkg/middleware"
	productDeliveryHttp "testTask/internal/pkg/product/delivery/http"
	productRepoPostgres "testTask/internal/pkg/product/repository/postgres"
	productRepoTarantool "testTask/internal/pkg/product/repository/tarantool"
	productUseCase "testTask/internal/pkg/product/usecase"
)

// todo read conf file or ENV
const (
	postgresConnStr         = "user=docker password=docker dbname=myService sslmode=disable port=5432 host=pg"
	postgresConnectionCount = 10

	tarantoolAddr = "tarantool:3301"
	tarantoolUser = "guest"

	goServerPortStr = ":8080"
)

func main() {

	// connect tarantool
	productRepoTarantoolObj, errConnect := productRepoTarantool.NewProductRepository(tarantoolAddr, tarantoolUser)
	if errConnect != nil {
		log.Fatal(errConnect)
	}
	defer func() {
		errClose := productRepoTarantoolObj.CloseProduct()
		if errClose != nil {
			log.Fatal(errClose)
		}
	}()
	log.Println("Tarantool connect successfully")

	// connect postgreSQL
	productRepoPostgresObj, errConnect := productRepoPostgres.NewProductRepository(postgresConnStr, postgresConnectionCount)
	if errConnect != nil {
		log.Fatal(errConnect)
	}
	defer func() {
		errClose := productRepoPostgresObj.CloseProduct()
		if errClose != nil {
			log.Fatal(errClose)
		}
	}()
	log.Println("PostgreSQL connect successfully")

	logger := logger.NewLogger("INFO")
	middlewareWithLogger := middleware.NewMiddlewareWithLogger(logger)

	productUseCaseObj := productUseCase.NewProductUseCase(productRepoPostgresObj, productRepoTarantoolObj, logger)
	productDeliveryObj := productDeliveryHttp.ProductHandler{ProductUseCase: productUseCaseObj}

	r := routing.New()

	r.Get("/products", func(c *routing.Context) error {
		productDeliveryObj.AddProducts(c)
		return nil
	})

	r.Get("/getProducts", func(c *routing.Context) error {
		productDeliveryObj.GetProducts(c)
		return nil
	})

	r.Get("/productsAsync", func(c *routing.Context) error {
		productDeliveryObj.AddProductsAsync(c)
		return nil
	})

	r.Get("/getStatus", func(c *routing.Context) error {
		productDeliveryObj.GetTaskStatus(c)
		return nil
	})

	routerWithMiddleware := middlewareWithLogger.LogMiddleware(r.HandleRequest)
	routerWithMiddleware = middlewareWithLogger.PanicMiddleware(routerWithMiddleware)

	log.Println("Start serving " + goServerPortStr)
	log.Fatal(fasthttp.ListenAndServe(goServerPortStr, routerWithMiddleware))
}
