package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type blobHandle struct {
	blob repository.BlobManager
}

func (b *blobHandle) GetRaw(c *gin.Context) {
	panic("unimplemented")
}

func (b *blobHandle) GetDecoded(c *gin.Context) {
	panic("unimplemented")
}

func (b *blobHandle) New(c *gin.Context) {
	panic("unimplemented")
}

func (b *blobHandle) Delete(c *gin.Context) {
	panic("unimplemented")
}
