package main

import (
	"context"
	"fmt"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/models/fake"
)

func main() {
	model := fake.New(fake.WithResponse("hello"))
	resp, err := model.Invoke(context.Background(), gopact.ModelRequest{
		Messages: []gopact.Message{gopact.UserMessage("say hello")},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp.Message.Parts[0].Text)
}
