package main

import "context"

type Executor interface {
	GetHandler(command string) (func(chatID int64, ctx context.Context, args string), bool)
}
