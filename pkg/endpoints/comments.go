package endpoints

import "github.com/whit-colm/itsc-4155-project/pkg/repository"

type commentHandle struct {
	book repository.BookManager
	comm repository.CommentManager
}

var ch commentHandle
