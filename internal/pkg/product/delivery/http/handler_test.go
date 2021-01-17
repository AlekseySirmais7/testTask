package http

import (
	"github.com/golang/mock/gomock"
	"testing"
)

func TestHandler_AddProducts(t *testing.T) {
	t.Helper()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

}
