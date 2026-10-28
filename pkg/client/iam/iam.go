// Package iam — клиент IAM-сервиса по контракту iam.proto (IamGrpc).
// В ЛР3 есть только заглушка Stub; живая реализация — ЛР5/ЛР6.
package iam

import "context"

type UserReply struct {
	ID       int64
	Email    string
	FullName string
	RoleName string
}

type Client interface {
	CheckPermission(ctx context.Context, userID int64, permission string) (bool, error)
	GetUser(ctx context.Context, userID int64) (*UserReply, error)
}
