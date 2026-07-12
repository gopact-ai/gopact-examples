package main

import (
	"context"
	"fmt"

	"github.com/gopact-ai/gopact/workflow"
)

func main() {
	wf := workflow.New[string, int]("length")
	count := wf.Node("count", func(_ context.Context, input string) (int, error) {
		return len(input), nil
	})
	wf.Entry(count)
	wf.Exit(count)

	out, err := wf.Invoke(context.Background(), "gopact")
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
