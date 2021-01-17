package product

import "testTask/internal/models"

type IUseCase interface {
	AddProducts(id int, xlsxFileUri string) (models.Statistic, error)
	GetProducts(sellerId int, offerId int, name string) ([]models.ProductWithSellerId, error)
	AddProductsAsync(sellerId int, xlsxFileUri string, taskId int)
	GetTaskStatus(taskId int) (string, error)
}
