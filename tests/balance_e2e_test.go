package tests

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestUsers(t *testing.T) {
	endpoint, teardown := provisionTestApp(t)
	defer teardown(t)
	res, err := http.Get(fmt.Sprintf("%s/balance/1", endpoint))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("response is: %s\n", b)
}
