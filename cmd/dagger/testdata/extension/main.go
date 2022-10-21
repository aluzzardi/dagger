package main

import (
	"context"

	"go-vanity-test.netlify.app/dagger/sdk/go/dagger"
)

type Test struct{}

func (Test) Test(ctx context.Context) (string, error) {
	return "hey", nil
}

func main() {
	dagger.Serve(Test{})
}
