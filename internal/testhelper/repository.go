package testhelper

import "github.com/whit-colm/itsc-4155-project/pkg/repository"

func Repository() *repository.Repository {
	r := repository.Repository{}
	return &r
}
