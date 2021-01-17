package usecase

import (
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"testTask/internal/models"
	"testTask/internal/pkg/constants"
	"testTask/internal/pkg/product"
	xlsxReader "testTask/internal/pkg/xlsx"
	"testTask/pkg/fileDownloader"
	"time"
)

type ProductUseCase struct {
	productRepo       product.IRepository
	productRepoStatus product.IRepositoryStatus
	logger            *zap.Logger
}

func NewProductUseCase(pr product.IRepository, prStatus product.IRepositoryStatus, logger *zap.Logger) product.IUseCase {
	return &ProductUseCase{
		productRepo:       pr,
		productRepoStatus: prStatus,
		logger:            logger,
	}
}

func (pu *ProductUseCase) AddProducts(sellerId int, xlsxFileUri string) (models.Statistic, error) {

	xlsxFileData, err := fileDownloader.DownloadFileByURI(xlsxFileUri)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProducts : DownloadFileByURI",
			zap.String("XlsxFileUri:", xlsxFileUri),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
		return models.Statistic{}, err
	}

	productsForAddOrUpdate, productsForDeleting, err := xlsxReader.GetProducts(xlsxFileData)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProducts : xlsxReader.GetProducts",
			zap.String("XlsxFileUri:", xlsxFileUri),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
		return models.Statistic{}, err
	}

	err = pu.checkValidProducts(&productsForAddOrUpdate)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProducts :checkValidProducts",
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
		return models.Statistic{}, err
	}

	err = pu.checkUniqProducts(&productsForAddOrUpdate, &productsForDeleting)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProducts : checkUniqProducts",
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
		return models.Statistic{}, err
	}

	addCount := 0
	deleteCount := 0
	updateCount := 0

	if len(productsForAddOrUpdate) > 0 {

		existProductIds, err := pu.productRepo.CheckProductExist(sellerId, &productsForAddOrUpdate)
		if err != nil {
			pu.logger.Error("ProductUseCase: AddProducts : CheckProductExist",
				zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
				zap.String("Error:", fmt.Sprintf("%v", err)),
			)
			return models.Statistic{}, err
		}

		productsForAdd, productsForUpdate := pu.separateAddAndUpdate(&productsForAddOrUpdate, &existProductIds)

		if len(productsForAdd) > 0 {
			addCount, err = pu.productRepo.AddProducts(sellerId, &productsForAdd)
			if err != nil {
				pu.logger.Error("ProductUseCase: AddProducts : AddProducts",
					zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
					zap.String("Error:", fmt.Sprintf("%v", err)),
				)
				return models.Statistic{}, err
			}
		}

		if len(productsForUpdate) > 0 {
			updateCount, err = pu.productRepo.UpdateProducts(sellerId, &productsForUpdate)
			if err != nil {
				pu.logger.Error("ProductUseCase: AddProducts : UpdateProducts",
					zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
					zap.String("Error:", fmt.Sprintf("%v", err)),
				)
				return models.Statistic{}, err
			}
		}
	}

	if len(productsForDeleting) > 0 {
		deleteCount, err = pu.productRepo.DeleteProducts(sellerId, &productsForDeleting)
		if err != nil {
			pu.logger.Error("ProductUseCase: AddProducts : DeleteProducts",
				zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
				zap.String("Error:", fmt.Sprintf("%v", err)),
			)
			return models.Statistic{}, err
		}
	}

	statistic := models.Statistic{
		AddCount:    addCount,
		DeleteCount: deleteCount,
		UpdateCount: updateCount,
	}
	return statistic, nil
}

func (pu *ProductUseCase) checkUniqProducts(productsAdd *[]models.Product, productsDelete *[]models.Product) error {
	for _, elementAdd := range *productsAdd {
		for _, elementDelete := range *productsDelete {
			if elementAdd.OfferId == elementDelete.OfferId {
				return errors.New("Same product for deleting and adding")
			}
		}
	}
	return nil
}

func (pu *ProductUseCase) separateAddAndUpdate(productsForAddOrUpdate *[]models.Product, existProductIds *[]int) (
	addProducts []models.Product,
	updateProducts []models.Product) {

	for _, elementAddOrUpdate := range *productsForAddOrUpdate {

		continueOuterLoop := false
		for _, updateId := range *existProductIds {
			if updateId == elementAddOrUpdate.OfferId {
				updateProducts = append(updateProducts, elementAddOrUpdate)
				continueOuterLoop = true
				break
			}
		}
		// not add updating product for adding
		if continueOuterLoop {
			continue
		}
		addProducts = append(addProducts, elementAddOrUpdate)
	}

	return addProducts, updateProducts
}

func (pu *ProductUseCase) GetProducts(sellerId int, offerId int, name string) ([]models.ProductWithSellerId, error) {
	//if sellerId == -1 && offerId == -1 && name == "" {
	//	return nil, errors.New("need set at least one of (sellerId, offerId, name)")
	//}
	// better return all products, if no filter

	pu.logger.Info("ProductUseCase:GetProducts",
		zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
		zap.String("OfferId:", fmt.Sprintf("%d", offerId)),
		zap.String("Name:", fmt.Sprintf("%s", name)),
	)

	products, err := pu.productRepo.GetProducts(sellerId, offerId, name)
	return products, err
}

func (pu *ProductUseCase) checkValidProducts(products *[]models.Product) error {

	for _, element := range *products {
		if element.Quantity < 0 {
			return errors.New("Bad value: Product quantity less 0")
		}
		if element.Price < 0 {
			return errors.New("Bad value: Product price less 0")
		}
		if element.OfferId < 0 {
			return errors.New("Bad value: Product offer_id less 0")
		}
		if element.Name == "" {
			return errors.New("Bad value: Product name empty")
		}
	}
	return nil
}

func (pu *ProductUseCase) AddProductsAsync(sellerId int, xlsxFileUri string, taskId int) {

	err := pu.productRepoStatus.CreateStatus(taskId, constants.StatusLoadXlsx)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProductsAsync : CreateStatus",
			zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
			zap.String("taskId:", fmt.Sprintf("%d", taskId)),
			zap.String("Status msg:", constants.StatusLoadXlsx),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
	}
	time.Sleep(5 * time.Second)

	xlsxFileData, err := fileDownloader.DownloadFileByURI(xlsxFileUri)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProductsAsync : DownloadFileByURI",
			zap.String("XlsxFileUri:", xlsxFileUri),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
		pu.productRepoStatus.SetStatus(taskId, fmt.Sprintf("Error: %v", err))
		return
	}

	err = pu.productRepoStatus.SetStatus(taskId, constants.StatusParseXlsx)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProductsAsync : SetStatus",
			zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
			zap.String("taskId:", fmt.Sprintf("%d", taskId)),
			zap.String("Status msg:", constants.StatusParseXlsx),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
	}
	time.Sleep(5 * time.Second)

	productsForAddOrUpdate, productsForDeleting, err := xlsxReader.GetProducts(xlsxFileData)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProductsAsync : xlsxReader.GetProducts",
			zap.String("XlsxFileUri:", xlsxFileUri),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
		pu.productRepoStatus.SetStatus(taskId, fmt.Sprintf("Error: %v", err))
		return
	}

	err = pu.productRepoStatus.SetStatus(taskId, constants.StatusCheckProductsFields)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProductsAsync : SetStatus",
			zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
			zap.String("taskId:", fmt.Sprintf("%d", taskId)),
			zap.String("Status msg:", constants.StatusCheckProductsFields),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
	}
	time.Sleep(5 * time.Second)

	err = pu.checkValidProducts(&productsForAddOrUpdate)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProductsAsync :checkValidProducts",
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
		pu.productRepoStatus.SetStatus(taskId, fmt.Sprintf("Error: %v", err))
		return
	}

	err = pu.productRepoStatus.SetStatus(taskId, constants.StatusCheckUniq)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProductsAsync : SetStatus",
			zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
			zap.String("taskId:", fmt.Sprintf("%d", taskId)),
			zap.String("Status msg:", constants.StatusCheckUniq),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
	}
	time.Sleep(5 * time.Second)

	err = pu.checkUniqProducts(&productsForAddOrUpdate, &productsForDeleting)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProductsAsync : checkUniqProducts",
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
		pu.productRepoStatus.SetStatus(taskId, fmt.Sprintf("Error: %v", err))
		return
	}

	addCount := 0
	deleteCount := 0
	updateCount := 0

	if len(productsForAddOrUpdate) > 0 {

		existProductIds, err := pu.productRepo.CheckProductExist(sellerId, &productsForAddOrUpdate)
		if err != nil {
			pu.logger.Error("ProductUseCase: AddProductsAsync : CheckProductExist",
				zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
				zap.String("Error:", fmt.Sprintf("%v", err)),
			)
			pu.productRepoStatus.SetStatus(taskId, fmt.Sprintf("Error: %v", err))
			return
		}

		productsForAdd, productsForUpdate := pu.separateAddAndUpdate(&productsForAddOrUpdate, &existProductIds)

		if len(productsForAdd) > 0 {

			err = pu.productRepoStatus.SetStatus(taskId, constants.StatusAddProducts)
			if err != nil {
				pu.logger.Error("ProductUseCase: AddProductsAsync : SetStatus",
					zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
					zap.String("taskId:", fmt.Sprintf("%d", taskId)),
					zap.String("Status msg:", constants.StatusAddProducts),
					zap.String("Error:", fmt.Sprintf("%v", err)),
				)
			}
			time.Sleep(5 * time.Second)

			addCount, err = pu.productRepo.AddProducts(sellerId, &productsForAdd)
			if err != nil {
				pu.logger.Error("ProductUseCase: AddProductsAsync : AddProducts",
					zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
					zap.String("Error:", fmt.Sprintf("%v", err)),
				)
				pu.productRepoStatus.SetStatus(taskId, fmt.Sprintf("Error: %v", err))
				return
			}
		}

		if len(productsForUpdate) > 0 {

			err = pu.productRepoStatus.SetStatus(taskId, constants.StatusUpdateProducts)
			if err != nil {
				pu.logger.Error("ProductUseCase: AddProductsAsync : SetStatus",
					zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
					zap.String("taskId:", fmt.Sprintf("%d", taskId)),
					zap.String("Status msg:", constants.StatusUpdateProducts),
					zap.String("Error:", fmt.Sprintf("%v", err)),
				)
			}
			time.Sleep(5 * time.Second)

			updateCount, err = pu.productRepo.UpdateProducts(sellerId, &productsForUpdate)
			if err != nil {
				pu.logger.Error("ProductUseCase: AddProductsAsync : UpdateProducts",
					zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
					zap.String("Error:", fmt.Sprintf("%v", err)),
				)
				pu.productRepoStatus.SetStatus(taskId, fmt.Sprintf("Error: %v", err))
				return
			}
		}
	}

	if len(productsForDeleting) > 0 {

		err = pu.productRepoStatus.SetStatus(taskId, constants.StatusDeleteProducts)
		if err != nil {
			pu.logger.Error("ProductUseCase: AddProductsAsync : SetStatus",
				zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
				zap.String("Status msg:", constants.StatusDeleteProducts),
				zap.String("taskId:", fmt.Sprintf("%d", taskId)),
				zap.String("Error:", fmt.Sprintf("%v", err)),
			)
		}
		time.Sleep(5 * time.Second)

		deleteCount, err = pu.productRepo.DeleteProducts(sellerId, &productsForDeleting)
		if err != nil {
			pu.logger.Error("ProductUseCase: AddProductsAsync : DeleteProducts",
				zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
				zap.String("Error:", fmt.Sprintf("%v", err)),
			)
			pu.productRepoStatus.SetStatus(taskId, fmt.Sprintf("Error: %v", err))
			return
		}
	}

	endStatus := fmt.Sprintf("Task end: { AddCount:%d  DeleteCount:%d  UpdateCount:%d",
		addCount,
		deleteCount,
		updateCount)

	err = pu.productRepoStatus.SetStatus(taskId, endStatus)
	if err != nil {
		pu.logger.Error("ProductUseCase: AddProductsAsync : SetStatus",
			zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
			zap.String("Status msg:", endStatus),
			zap.String("taskId:", fmt.Sprintf("%d", taskId)),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
	}
	pu.logger.Info("ProductUseCase: AddProductsAsync : AddProductsAsync : Set end status",
		zap.String("SellerId:", fmt.Sprintf("%d", sellerId)),
		zap.String("Status msg:", endStatus),
		zap.String("taskId:", fmt.Sprintf("%d", taskId)),
		zap.String("Error:", fmt.Sprintf("%v", err)),
	)
}

func (pu *ProductUseCase) GetTaskStatus(taskId int) (string, error) {
	status, err := pu.productRepoStatus.GetStatus(taskId)
	if err != nil {
		pu.logger.Error("ProductUseCase: GetTaskStatus : GetStatus",
			zap.String("taskId:", fmt.Sprintf("%d", taskId)),
			zap.String("Error:", fmt.Sprintf("%v", err)),
		)
	}
	return status, err
}
