package model_test

import (
	"encoding/json"
	"testing"

	"github.com/whit-colm/itsc-4155-project/internal/testhelper/dummyvalues"
)

func TestJSONMarshal(t *testing.T) {
	for _, v := range dummyvalues.ExampleBooks {
		_, err := json.Marshal(v)
		if err != nil {
			t.Errorf("unable to marshal individual book %v: %s", v, err)
		}
	}

	_, err := json.Marshal(dummyvalues.ExampleBooks)
	if err != nil {
		t.Errorf("unable to marshal into slice: %s", err)
	}
}

/*func TestJSONUnmarshal(t *testing.T) {
	individualBooksJson := [][]byte{[]byte(`{"id":"01953a93-21e7-73da-8a27-fc22aa66a95e","isbns":[{"type":"isbn10","value":"0141439602"},{"type":"isbn13","value":"9780141439600"}],"title":"A Tale of Two Cities","author":"Charles Dickens","published":"1859-11-26"}`),
		[]byte(`{"id":"01953a93-21e7-73dd-84c7-a4df7fadac5a","isbns":[{"type":"isbn10","value":"0156012197"},{"type":"isbn13","value":"9780156012195"}],"title":"The Little Prince","author":"Antoine de Saint-Exupéry","published":"1943-04-00"}`),
		[]byte(`{"id":"01953a93-21e7-73de-84dd-33f54daba1ec","isbns":[{"type":"isbn10","value":"0062315005"},{"type":"isbn13","value":"9780061122415"}],"title":"The Alchemist","author":"Paulo Coelho","published":"1988-00-00"}`)}

	listBooksJson := []byte([{"id":"01953a9c-55a1-7c04-b0f1-71c51c1cc8c5","isbns":[{"type":"isbn10","value":"0141439602"},{"type":"isbn13","value":"9780141439600"}],"title":"A Tale of Two Cities","author":"Charles Dickens","published":"1859-11-26"},{"id":"01953a9c-55a1-7c07-9576-05c0deb47fe5","isbns":[{"type":"isbn10","value":"0156012197"},{"type":"isbn13","value":"9780156012195"}],"title":"The Little Prince","author":"Antoine de Saint-Exupéry","published":"1943-04-00"},{"id":"01953a9c-55a1-7c08-be47-6530a96888ae","isbns":[{"type":"isbn10","value":"0062315005"},{"type":"isbn13","value":"9780061122415"}],"title":"The Alchemist","author":"Paulo Coelho","published":"1988-00-00"}])
}*/
