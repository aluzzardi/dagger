package main

import (
	"context"
	"os"

	"go-vanity-test.netlify.app/dagger/sdk/go/dagger"
)

type Test struct{}

func (Test) TestMount(ctx context.Context, in dagger.DirectoryID) (string, error) {
	bytes, err := os.ReadFile("/mnt/in/foo")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func main() {
	dagger.Serve(Test{})
}
