package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type commentHandle struct {
	book repository.BookManager
	comm repository.CommentManager
}

var ch commentHandle

func (ch *commentHandle) BookReviews(c *gin.Context) {
	panic("unimplemented")
}

func (ch *commentHandle) Get(c *gin.Context) {
	panic("unimplemented")
}

func (ch *commentHandle) Post(c *gin.Context) {
	panic("unimplemented")
}

func (ch *commentHandle) Edit(c *gin.Context) {
	panic("unimplemented")
}

func (ch *commentHandle) Delete(c *gin.Context) {
	panic("unimplemented")
}
