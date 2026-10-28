package resolver

import (
	"context"
	"time"

	appointmentdomain "github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	catalogdomain "github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	notificationdomain "github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	paymentdomain "github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
)

// Интерфейсы того, что gateway просит у сервисов. В ЛР3 их реализуют сами
// сервисы в том же процессе (container передаёт *service.XxxService), в ЛР6 —
// gRPC-клиенты с теми же методами. Резолверы только переводят аргументы и
// ответы, бизнес-логика остаётся в сервисах.

type CatalogService interface {
	ListSpecialties(ctx context.Context) ([]catalogdomain.Specialty, error)
	ListDoctors(ctx context.Context, specialtyID *int64) ([]catalogdomain.Doctor, error)
	GetPatientByUserID(ctx context.Context, userID int64) (*catalogdomain.Patient, error)
	ListMedicalServices(ctx context.Context) ([]catalogdomain.MedicalService, error)
	ListMedications(ctx context.Context) ([]catalogdomain.Medication, error)
}

type AppointmentService interface {
	FreeSlots(ctx context.Context, doctorID int64, date string) ([]appointmentdomain.Slot, error)
	GetAppointment(ctx context.Context, id int64) (*appointmentdomain.Appointment, error)
	ListPatientAppointments(ctx context.Context, patientID int64, status *appointmentdomain.Status) ([]appointmentdomain.Appointment, error)
}

type PaymentService interface {
	ListPayments(ctx context.Context, patientID *int64) ([]paymentdomain.Payment, error)
}

type NotificationService interface {
	ListUserNotifications(ctx context.Context, userID int64, status *notificationdomain.Status) ([]notificationdomain.Notification, error)
}

type Resolver struct {
	catalog       CatalogService
	appointments  AppointmentService
	payments      PaymentService
	notifications NotificationService
	location      *time.Location
	// currentUserID — ВРЕМЕННО до ЛР5: «текущий пользователь» для my*-запросов
	// берётся из конфига (GATEWAY_TEST_USER_ID), а не из JWT.
	currentUserID int64
}

func New(
	catalog CatalogService,
	appointments AppointmentService,
	payments PaymentService,
	notifications NotificationService,
	location *time.Location,
	currentUserID int64,
) *Resolver {
	return &Resolver{
		catalog:       catalog,
		appointments:  appointments,
		payments:      payments,
		notifications: notifications,
		location:      location,
		currentUserID: currentUserID,
	}
}
