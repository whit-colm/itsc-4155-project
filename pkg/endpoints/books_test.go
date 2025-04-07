package endpoints

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/internal/testhelper"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

func TestMain(m *testing.M) {
	r := gin.Default()
	ds := repository.Repository{
		Store: &testhelper.TestingStoreManager{},
		Book:  &testhelper.TestingBookRepository{},
	}
	Configure(r, &ds, nil)
	os.Exit(m.Run())
}
