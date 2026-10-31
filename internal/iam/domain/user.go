package domain

import "time"

type User struct {
	ID        int64
	RoleID    int64
	RoleName  string
	Email     string
	FullName  string
	Phone     *string
	CreatedAt time.Time
}
