package endpoints

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type blobHandle struct {
	blob repository.BlobManager
}

var lh blobHandle

func (b *blobHandle) GetRaw(c *gin.Context) (int, string, error) {
	const errorCaller string = "get  blob"
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return http.StatusBadRequest,
			"Unable to parse UUID",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	o, err := lh.blob.GetByID(c.Request.Context(), id)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}

	if bb, err := io.ReadAll(o.Content); err != nil {
		return http.StatusInternalServerError,
			"Error reading blob into response",
			fmt.Errorf("%v: %w", errorCaller, err)
	} else {

		for k, v := range o.Metadata {
			switch k {
			// TODO: update in future
			case "last-modified":
				c.Header(k, v)
			default:
				continue
			}
		}
		c.Data(http.StatusOK, o.Metadata["content-type"], bb)
	}
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
