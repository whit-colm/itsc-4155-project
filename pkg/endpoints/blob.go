package endpoints

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type blobHandle struct {
	blob repository.BlobManager
}

var lh blobHandle

func (b *blobHandle) get(c *gin.Context) (*model.Blob, int, string, error) {
	const errorCaller string = "get blob"
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return &model.Blob{},
			http.StatusBadRequest,
			"Unable to parse UUID",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	if blob, err := lh.blob.GetByID(c.Request.Context(), id); err != nil {
		h, s, e := wrapDatastoreError(errorCaller, err)
		return &model.Blob{}, h, s, e
	} else {
		return blob, http.StatusOK, "", nil
	}
}

func (b *blobHandle) GetRaw(c *gin.Context) (int, string, error) {
	const errorCaller string = "get raw blob"
	o, h, s, e := b.get(c)
	if e != nil {
		return h, s, e
	}
	if bb, err := io.ReadAll(o.Content); err != nil {
		return http.StatusInternalServerError,
			"Error reading blob into response",
			fmt.Errorf(errorCaller, err)
	} else {
		for k, v := range o.Metadata {
			switch k {
			// TODO: update in future
			case "last-modified":
				c.Request.Response.Header.Add(k, v)
			}
		}
		c.Data(http.StatusOK, o.Metadata["content-type"], bb)
	}
}

func (b *blobHandle) GetAsJSON(c *gin.Context) (int, string, error) {
	const errorCaller string = "get blob JSON"
	o, h, s, e := b.get(c)
	if e != nil {
		return h, s, e
	}
	c.JSON(http.StatusOK, o)
	return http.StatusOK, "", nil
}

func (b *blobHandle) New(c *gin.Context) (int, string, error) {
	const errorCaller string = "create blob"
	panic("unimplemented")
}

func (b *blobHandle) Delete(c *gin.Context) (int, string, error) {
	const errorCaller string = "delete blob"
	panic("unimplemented")
}
