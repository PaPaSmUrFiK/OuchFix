package iam

import (
	"context"
	"fmt"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

// Stub — фейковый IamGrpc для ЛР3: пользователи и права как в seed миграций.
type Stub struct {
	users       map[int64]UserReply
	permissions map[string]map[string]struct{}
}

var _ Client = (*Stub)(nil)

func NewStub() *Stub {
	return &Stub{
		users: map[int64]UserReply{
			1: {ID: 1, Email: "admin@ouchfix.by", FullName: "Администратор Центра", RoleName: "admin"},
			2: {ID: 2, Email: "ivanov@ouchfix.by", FullName: "Иванов Пётр Сергеевич", RoleName: "doctor"},
			3: {ID: 3, Email: "petrova@ouchfix.by", FullName: "Петрова Анна Игоревна", RoleName: "doctor"},
			4: {ID: 4, Email: "sidorov@ouchfix.by", FullName: "Сидоров Олег Викторович", RoleName: "doctor"},
			5: {ID: 5, Email: "kozlov@mail.by", FullName: "Козлов Артём Андреевич", RoleName: "patient"},
			6: {ID: 6, Email: "smirnova@mail.by", FullName: "Смирнова Мария Олеговна", RoleName: "patient"},
			7: {ID: 7, Email: "novik@mail.by", FullName: "Новик Денис Павлович", RoleName: "patient"},
		},
		permissions: map[string]map[string]struct{}{
			"patient": set("appointment:create", "appointment:read", "payment:create"),
			"doctor":  set("appointment:read", "appointment:complete"),
			"admin": set("appointment:create", "appointment:read", "schedule:manage",
				"medication:manage", "payment:create"),
		},
	}
}

func (s *Stub) GetUser(ctx context.Context, userID int64) (*UserReply, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	user, ok := s.users[userID]
	if !ok {
		return nil, fmt.Errorf("iam stub: user %d: %w", userID, pkgerrors.ErrNotFound)
	}
	return &user, nil
}

func (s *Stub) CheckPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return false, err
	}
	_, ok := s.permissions[user.RoleName][permission]
	return ok, nil
}

func set(codes ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		out[c] = struct{}{}
	}
	return out
}
