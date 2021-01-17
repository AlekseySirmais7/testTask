package tarantool

import (
	tarantool "github.com/tarantool/go-tarantool"
	"testTask/internal/pkg/product"
)

type ProductRepository struct {
	connectTarantool *tarantool.Connection
}

func NewProductRepository(tarantoolAddr string, tarantoolUser string) (product.IRepositoryStatus, error) {
	opts := tarantool.Opts{User: tarantoolUser}
	tarantoolConnect, err := tarantool.Connect(tarantoolAddr, opts)
	if err != nil {
		return nil, err
	}
	_, errPing := tarantoolConnect.Ping()
	if errPing != nil {
		return nil, errPing
	}
	return &ProductRepository{connectTarantool: tarantoolConnect}, nil
}

func (prStatus *ProductRepository) CloseProduct() error {
	errClose := prStatus.connectTarantool.Close()
	if errClose != nil {
		return errClose
	}
	return nil
}

func (prStatus *ProductRepository) CreateStatus(taskId int, status string) error {
	var params []interface{}
	params = append(params, taskId, status)
	_, err := prStatus.connectTarantool.Insert("statuses", params)
	return err
}

func (prStatus *ProductRepository) SetStatus(taskId int, status string) error {
	var params []interface{}
	params = append(params, taskId, status)
	_, err := prStatus.connectTarantool.Replace("statuses", params)
	return err
}

func (prStatus *ProductRepository) GetStatus(taskId int) (string, error) {
	var params []interface{}
	params = append(params, taskId)
	resp, err := prStatus.connectTarantool.Select("statuses", "primary", 0, 1, tarantool.IterEq, params)
	if err != nil {
		return "", err
	}
	statusData := resp.Tuples()
	if len(statusData) != 1 {
		return "no status for this taskId", err
	}
	statusStr := statusData[0][1].(string)
	return statusStr, err
}
