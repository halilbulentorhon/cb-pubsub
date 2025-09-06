package model

import "time"

type PubSubDoc[T any] struct {
	Messages     []T   `json:"messages"`
	CreationDate int64 `json:"creationDate"`
}

func CreatePubSubDoc[T any]() PubSubDoc[T] {
	currentTimestamp := time.Now().Unix()
	return PubSubDoc[T]{
		CreationDate: currentTimestamp,
		Messages:     CreateEmptyMessages[T](),
	}
}

func CreateEmptyMessages[T any]() []T {
	return make([]T, 0)
}
