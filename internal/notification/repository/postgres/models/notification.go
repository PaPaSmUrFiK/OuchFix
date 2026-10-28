package models

import (
	"database/sql"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
)

type Notification struct {
	ID          int64          `db:"id"`
	UserID      int64          `db:"user_id"`
	Channel     string         `db:"channel"`
	Title       string         `db:"title"`
	Body        string         `db:"body"`
	Status      string         `db:"status"`
	RelatedType sql.NullString `db:"related_type"`
	RelatedID   sql.NullInt64  `db:"related_id"`
	CreatedAt   time.Time      `db:"created_at"`
	SentAt      sql.NullTime   `db:"sent_at"`
}

func (m Notification) ToDomain() domain.Notification {
	n := domain.Notification{
		ID:        m.ID,
		UserID:    m.UserID,
		Channel:   m.Channel,
		Title:     m.Title,
		Body:      m.Body,
		Status:    domain.Status(m.Status),
		CreatedAt: m.CreatedAt,
	}
	if m.RelatedType.Valid {
		v := m.RelatedType.String
		n.RelatedType = &v
	}
	if m.RelatedID.Valid {
		v := m.RelatedID.Int64
		n.RelatedID = &v
	}
	if m.SentAt.Valid {
		v := m.SentAt.Time
		n.SentAt = &v
	}
	return n
}

func ToDomainList(rows []Notification) []domain.Notification {
	out := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToDomain())
	}
	return out
}
