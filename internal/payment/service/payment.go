package service

import (
	"context"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type PaymentRepository interface {
	List(ctx context.Context, patientID *int64) ([]domain.Payment, error)
}

type PaymentService struct {
	repo PaymentRepository
}

func NewPaymentService(repo PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

// ListPayments возвращает платежи пациента; patientID == nil — все платежи.
func (s *PaymentService) ListPayments(ctx context.Context, patientID *int64) ([]domain.Payment, error) {
	if patientID != nil {
		if err := pkgerrors.ValidateID("patientId", *patientID); err != nil {
			return nil, err
		}
	}

	return s.repo.List(ctx, patientID)
}
