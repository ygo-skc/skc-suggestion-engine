package testing

import (
	"context"
	"os"
	"path"
	"runtime"
)

var TestContext context.Context

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	TestContext = context.Background()
}
