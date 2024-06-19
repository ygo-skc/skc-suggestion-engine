package testing

import (
	"context"
	"log/slog"
	"os"
	"path"
	"runtime"

	"github.com/ygo-skc/skc-suggestion-engine/util"
)

var TestContext context.Context

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}

	TestContext = context.WithValue(context.Background(), util.Logger, slog.With())
}
