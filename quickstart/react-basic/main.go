package main

import (
	"context"
	"fmt"

	"github.com/gopact-ai/gopact"
	"github.com/gopact-ai/gopact-ext/agents/react"
	"github.com/gopact-ai/gopact-ext/models/fake"
	"github.com/gopact-ai/gopact/agent"
)

func main() {
	target, err := react.New(
		agent.Identity{
			Name:        "quickstart",
			Description: "demonstrates the react agent",
			Version:     "v1",
		},
		fake.New(fake.WithResponse("done")),
	)
	if err != nil {
		panic(err)
	}
	response, err := target.Invoke(context.Background(), agent.Request{
		Messages: []gopact.Message{gopact.UserMessage("finish")},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(response.Message.Parts[0].Text)
}
