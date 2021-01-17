package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
	"strconv"
	"testTask/internal/models"
	"testTask/internal/pkg/product"
)

type ProductRepository struct {
	connectPg *sql.DB
}

func NewProductRepository(postgresConnStr string, ConnectionCount int) (product.IRepository, error) {
	postgresDB, err := sql.Open("postgres", postgresConnStr)
	if err != nil {
		return nil, err
	}
	postgresDB.SetMaxOpenConns(ConnectionCount)

	errPing := postgresDB.Ping()
	if errPing != nil {
		return nil, errPing
	}

	return &ProductRepository{connectPg: postgresDB}, nil
}

func (pc *ProductRepository) CloseProduct() error {
	errClose := pc.connectPg.Close()
	if errClose != nil {
		return errClose
	}
	return nil
}

func (pc *ProductRepository) AddProducts(sellerId int, products *[]models.Product) (int, error) {

	const maxProductsPerRequest = 1000

	if len(*products) <= maxProductsPerRequest {
		insertCount, err := pc.AddProductsRequest(sellerId, products)
		return insertCount, err
	}

	currentPartSize := maxProductsPerRequest
	alreadyInsertedCount := 0
	insertCount := 0

	for isContinue := true; isContinue; isContinue = alreadyInsertedCount < len(*products) {

		productPart := make([]models.Product, currentPartSize)

		copy(productPart, (*products)[alreadyInsertedCount:alreadyInsertedCount+currentPartSize])

		insertPartCount, err := pc.AddProductsRequest(sellerId, &productPart)
		if err != nil {
			return 0, err
		}

		alreadyInsertedCount += currentPartSize
		insertCount += insertPartCount

		if len(*products)-alreadyInsertedCount >= maxProductsPerRequest {
			currentPartSize = maxProductsPerRequest
		} else {
			currentPartSize = (len(*products) - alreadyInsertedCount) % maxProductsPerRequest
		}
	}

	return insertCount, nil
}

func (pc *ProductRepository) AddProductsRequest(sellerId int, products *[]models.Product) (int, error) {

	insertProductsSQL := "INSERT INTO Products (seller_id, offer_id, name, price, quantity, available) VALUES \n"

	productValuesSQL := ""
	sqlQueryValues := []interface{}{}

	// collect product
	for i, element := range *products {

		startIndex := i * 6

		productValuesSQL += `( $` + strconv.Itoa(startIndex+1) +
			`, $` + strconv.Itoa(startIndex+2) +
			`, $` + strconv.Itoa(startIndex+3) +
			`, $` + strconv.Itoa(startIndex+4) +
			`, $` + strconv.Itoa(startIndex+5) +
			`, $` + strconv.Itoa(startIndex+6) + ` ) `

		if i+1 != len(*products) {
			productValuesSQL += ", \n"
		}

		sqlQueryValues = append(sqlQueryValues,
			sellerId,
			element.OfferId,
			element.Name,
			element.Price,
			element.Quantity,
			element.Available)
	}

	insertProductsSQL += productValuesSQL + " ; "

	stmt, err := pc.connectPg.Prepare(insertProductsSQL)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	resultSelect, err := stmt.Exec(sqlQueryValues...)
	if err != nil {
		return 0, err
	}

	addCount, err := resultSelect.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(addCount), nil
}

func (pc *ProductRepository) UpdateProducts(sellerId int, products *[]models.Product) (int, error) {

	const maxProductsPerRequest = 1000

	if len(*products) <= maxProductsPerRequest {
		updateCount, err := pc.UpdateProductsRequest(sellerId, products)
		return updateCount, err
	}

	currentPartSize := maxProductsPerRequest
	alreadyUpdateCount := 0
	updateCount := 0

	for isContinue := true; isContinue; isContinue = alreadyUpdateCount < len(*products) {

		productPart := make([]models.Product, currentPartSize)

		copy(productPart, (*products)[alreadyUpdateCount:alreadyUpdateCount+currentPartSize])

		updatePartCount, err := pc.UpdateProductsRequest(sellerId, &productPart)
		if err != nil {
			return 0, err
		}

		alreadyUpdateCount += currentPartSize
		updateCount += updatePartCount

		if len(*products)-alreadyUpdateCount >= maxProductsPerRequest {
			currentPartSize = maxProductsPerRequest
		} else {
			currentPartSize = (len(*products) - alreadyUpdateCount) % maxProductsPerRequest
		}
	}
	return updateCount, nil
}

func (pc *ProductRepository) UpdateProductsRequest(sellerId int, products *[]models.Product) (int, error) {

	updateProductsSQLFirstPart := `UPDATE Products AS p SET
							name = c.name::text,
							price = c.price::int,
							quantity = c.quantity::int,
							available = c.available::bool
							FROM ( VALUES `

	updateProductsSQLSecondPart := "" // values from products will be here

	updateProductsSQLThirdPart := ` ) AS c (offer_id, name, price, quantity, available)
									WHERE c.offer_id::int = p.offer_id::int AND p.seller_id = $` +
		strconv.Itoa(len(*products)*5+1) + ";"
	// 5 product's fields and + 1 for  position seller_id

	sqlQueryValues := []interface{}{}

	// collect updating product
	for i, element := range *products {

		startIndex := i * 5

		updateProductsSQLSecondPart += ` ( $` + strconv.Itoa(startIndex+1) +
			`, $` + strconv.Itoa(startIndex+2) +
			`, $` + strconv.Itoa(startIndex+3) +
			`, $` + strconv.Itoa(startIndex+4) +
			`, $` + strconv.Itoa(startIndex+5) + ` ) `

		if i+1 != len(*products) {
			updateProductsSQLSecondPart += ", \n"
		}

		sqlQueryValues = append(sqlQueryValues,
			element.OfferId,
			element.Name,
			element.Price,
			element.Quantity,
			element.Available)
	}

	// add sellerId for $ end of updateProductsSQLThirdPart
	sqlQueryValues = append(sqlQueryValues, sellerId)

	updateProductsSQL := updateProductsSQLFirstPart + updateProductsSQLSecondPart + updateProductsSQLThirdPart

	stmt, err := pc.connectPg.Prepare(updateProductsSQL)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	resultUpdate, err := stmt.Exec(sqlQueryValues...)
	if err != nil {
		return 0, err
	}

	updateCount, err := resultUpdate.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(updateCount), nil
}

func (pc *ProductRepository) CheckProductExist(sellerId int, products *[]models.Product) ([]int, error) {

	// main part
	getProductsBySellerIdAndOfferIdSQL := "SELECT offer_id FROM Products WHERE seller_id = $1 AND ( offer_id IN ("

	// part with sellerId and []products.OfferId
	productOfferIdSQL := ""
	sqlQueryValues := []interface{}{}

	sqlQueryValues = append(sqlQueryValues, sellerId)

	// collect ids
	for i, element := range *products {
		sqlQueryValues = append(sqlQueryValues, element.OfferId)
		productOfferIdSQL += ` $` + strconv.Itoa(i+2)
		if i+1 != len(*products) {
			productOfferIdSQL += `, `
		}
	}

	getProductsBySellerIdAndOfferIdSQL += productOfferIdSQL + " ) );"

	stmt, err := pc.connectPg.Prepare(getProductsBySellerIdAndOfferIdSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	resultSelect, err := stmt.Query(sqlQueryValues...)
	if err != nil {
		return nil, err
	}

	existProductsId := []int{}
	for resultSelect.Next() {
		var oneId int
		errScan := resultSelect.Scan(&oneId)
		if errScan != nil {
			return nil, errScan
		}
		existProductsId = append(existProductsId, oneId)
	}
	return existProductsId, nil
}

func (pc *ProductRepository) DeleteProducts(sellerId int, products *[]models.Product) (int, error) {

	deleteProductsSQL := "DELETE FROM Products WHERE seller_id = $1 AND offer_id IN ( "

	productValuesSQL := ""
	sqlQueryValues := []interface{}{}

	sqlQueryValues = append(sqlQueryValues, sellerId)

	// collect product.offerId
	for i, element := range *products {
		productValuesSQL += ` $` + strconv.Itoa(i+2)
		if i+1 != len(*products) {
			productValuesSQL += ", "
		}
		sqlQueryValues = append(sqlQueryValues, element.OfferId)
	}

	deleteProductsSQL += productValuesSQL + " ); "

	stmt, err := pc.connectPg.Prepare(deleteProductsSQL)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	resultDelete, err := stmt.Exec(sqlQueryValues...)
	if err != nil {
		return 0, err
	}

	deleteCount, err := resultDelete.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(deleteCount), nil
}

func (pc *ProductRepository) GetProducts(sellerId int, offerId int, name string) ([]models.ProductWithSellerId, error) {

	selectProductsSQL := `SELECT seller_id, offer_id, name, price, quantity, available
                          FROM Products `

	queryParameterIter := 1
	sqlQueryValues := []interface{}{}

	if sellerId != -1 {
		selectProductsSQL += `WHERE seller_id = $` + strconv.Itoa(queryParameterIter)
		queryParameterIter++
		sqlQueryValues = append(sqlQueryValues, sellerId)
	}

	if offerId != -1 {
		if queryParameterIter != 1 {
			selectProductsSQL += " AND "
		} else {
			selectProductsSQL += " WHERE "
		}
		selectProductsSQL += ` offer_id = $` + strconv.Itoa(queryParameterIter)
		queryParameterIter++
		sqlQueryValues = append(sqlQueryValues, offerId)
	}
	if name != "" {
		if queryParameterIter != 1 {
			selectProductsSQL += " AND "
		} else {
			selectProductsSQL += " WHERE "
		}
		//selectProductsSQL += ` to_tsvector("name") @@ plainto_tsquery($` + strconv.Itoa(queryParameterIter) + ` )`
		selectProductsSQL += ` LOWER(name) LIKE LOWER('%'::text || $` + strconv.Itoa(queryParameterIter) + `::text || '%'::text) `
		sqlQueryValues = append(sqlQueryValues, name)
	}

	stmt, err := pc.connectPg.Prepare(selectProductsSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	selectedProducts, err := stmt.Query(sqlQueryValues...)
	if err != nil {
		return nil, err
	}

	var products = []models.ProductWithSellerId{}
	for selectedProducts.Next() {
		var oneProduct models.ProductWithSellerId
		err := selectedProducts.Scan(&oneProduct.SellerId,
			&oneProduct.OfferId,
			&oneProduct.Name,
			&oneProduct.Price,
			&oneProduct.Quantity,
			&oneProduct.Available)
		if err != nil {
			return nil, err
		}
		products = append(products, oneProduct)
	}
	return products, nil
}
