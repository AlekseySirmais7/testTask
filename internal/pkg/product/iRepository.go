package product

import "testTask/internal/models"

type IRepository interface {
	AddProducts(sellerId int, products *[]models.Product) (int, error)
	UpdateProducts(sellerId int, products *[]models.Product) (int, error)
	CheckProductExist(sellerId int, products *[]models.Product) ([]int, error)
	DeleteProducts(sellerId int, products *[]models.Product) (int, error)
	GetProducts(sellerId int, offerId int, name string) ([]models.ProductWithSellerId, error)
	CloseProduct() error
}
