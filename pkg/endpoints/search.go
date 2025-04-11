package endpoints

import (
	"github.com/gin-gonic/gin"

	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type searchHandle[S comparable] struct {
	book repository.BookManager[S]
	athr repository.AuthorManager[S]
	comm repository.CommentManager[S]
}

func (h searchHandle[S]) Search(c *gin.Context) {
	panic("unimplemented")
}
