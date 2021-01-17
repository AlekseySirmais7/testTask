package xlsx

import (
	"errors"
	"github.com/tealeg/xlsx"
	"testTask/internal/models"
)

func GetProducts(xlsxFileData []byte) ([]models.Product, []models.Product, error) {

	xlsxDoc, err := xlsx.OpenBinary(xlsxFileData)
	if err != nil {
		return nil, nil, errors.New("err xlsx.OpenBinary(): " + err.Error())
	}

	productsForAdding := []models.Product{}
	productsForDeleting := []models.Product{}

	for i := startDataConst; i <= xlsxDoc.Sheets[0].MaxRow; i++ {

		oneProduct := models.Product{}

		theCheckCell, err := xlsxDoc.Sheets[0].Cell(i, 0)
		if err != nil {
			return nil, nil, errors.New("err Cell(" + string(i) + ",0) " + err.Error())
		}

		// empty line -> break,
		// MaxRow ~ 1000, if we have only 5 products
		if theCheckCell.String() == "" {
			break
		}

		for j := 0; j <= colCountConst; j++ {

			theCell, err := xlsxDoc.Sheets[0].Cell(i, j)
			if err != nil {
				return nil, nil, errors.New("err Cell(" + string(i) + "," + string(j) + ") " + err.Error())
			}

			switch j {
			case offerIdCol:
				offerId, errParseInt := theCell.Int()
				if errParseInt != nil {
					return nil, nil, errors.New("err Cell.Int(): " + errParseInt.Error())
				}
				oneProduct.OfferId = offerId

			case nameCol:
				oneProduct.Name = theCell.String()

			case priceCol:
				price, errParseFloat := theCell.Int()
				if errParseFloat != nil {
					return nil, nil, errors.New("err Cell.Float(): " + errParseFloat.Error())
				}
				oneProduct.Price = price

			case quantityCol:
				quantity, errParseInt := theCell.Int()
				if errParseInt != nil {
					return nil, nil, errors.New("err Cell.Int(): " + errParseInt.Error())
				}
				oneProduct.Quantity = quantity

			case availableCol:
				oneProduct.Available = theCell.Bool()

			default:
				continue
			}
		}
		if oneProduct.Available {
			productsForAdding = append(productsForAdding, oneProduct)
		} else {
			productsForDeleting = append(productsForDeleting, oneProduct)
		}

	}
	return productsForAdding, productsForDeleting, nil
}
