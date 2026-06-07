package application

import (
	"context"
	"testing"
	"time"

	billingdomain "github.com/eduaccess/eduaccess-api/internal/billing/domain"
	schooldomain "github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

type fakePaymentRepo struct {
	created        *billingdomain.PaymentTransaction
	updated        *billingdomain.PaymentTransaction
	findByID       *billingdomain.PaymentTransaction
	findByOrderID  *billingdomain.PaymentTransaction
	findPending    *billingdomain.PaymentTransaction
	findPendingErr error
	findByIDErr    error
	findByOrderErr error
}

func (f *fakePaymentRepo) Create(_ context.Context, payment *billingdomain.PaymentTransaction) error {
	f.created = payment
	return nil
}

func (f *fakePaymentRepo) FindByID(context.Context, uuid.UUID) (*billingdomain.PaymentTransaction, error) {
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}
	return f.findByID, nil
}

func (f *fakePaymentRepo) FindByProviderOrderID(context.Context, string) (*billingdomain.PaymentTransaction, error) {
	if f.findByOrderErr != nil {
		return nil, f.findByOrderErr
	}
	return f.findByOrderID, nil
}

func (f *fakePaymentRepo) FindLatestPendingBySchool(context.Context, uuid.UUID) (*billingdomain.PaymentTransaction, error) {
	if f.findPendingErr != nil {
		return nil, f.findPendingErr
	}
	return f.findPending, nil
}

func (f *fakePaymentRepo) Update(_ context.Context, payment *billingdomain.PaymentTransaction) error {
	f.updated = payment
	return nil
}

type fakeBillingSchoolRepo struct {
	school               *schooldomain.School
	plan                 *schooldomain.Plan
	activeSubscription   *schooldomain.Subscription
	replacedSubscription *schooldomain.Subscription
}

func (f *fakeBillingSchoolRepo) FindByID(context.Context, uuid.UUID) (*schooldomain.School, error) {
	return f.school, nil
}

func (f *fakeBillingSchoolRepo) FindPlanByID(context.Context, uuid.UUID) (*schooldomain.Plan, error) {
	return f.plan, nil
}

func (f *fakeBillingSchoolRepo) FindActiveSubscription(context.Context, uuid.UUID) (*schooldomain.Subscription, error) {
	if f.activeSubscription == nil {
		return nil, apperror.New(apperror.ErrNotFound, "no active subscription found")
	}
	return f.activeSubscription, nil
}

func (f *fakeBillingSchoolRepo) ReplaceSubscription(_ context.Context, sub *schooldomain.Subscription) error {
	f.replacedSubscription = sub
	return nil
}

type fakeStudentCounter struct {
	total int64
}

func (f *fakeStudentCounter) CountActiveStudents(context.Context, uuid.UUID) (int64, error) {
	return f.total, nil
}

type fakeGateway struct {
	session       *GatewayCheckoutSession
	status        *GatewayTransactionStatus
	statusErr     error
	verifyResult  bool
	receivedInput *GatewayCreateCheckoutInput
}

func (f *fakeGateway) CreateCheckout(_ context.Context, input GatewayCreateCheckoutInput) (*GatewayCheckoutSession, error) {
	f.receivedInput = &input
	return f.session, nil
}

func (f *fakeGateway) GetTransactionStatus(context.Context, string) (*GatewayTransactionStatus, error) {
	return f.status, f.statusErr
}

func (f *fakeGateway) VerifySignature(orderID, statusCode, grossAmount, signature string) bool {
	return f.verifyResult
}

func TestCreateCheckoutHandler_CreatesPendingPaymentForUpgrade(t *testing.T) {
	schoolID := uuid.New()
	planID := uuid.New()
	userID := uuid.New()
	repo := &fakePaymentRepo{findPendingErr: apperror.New(apperror.ErrNotFound, "not found")}
	schools := &fakeBillingSchoolRepo{
		school: &schooldomain.School{ID: schoolID, Name: "SMK Nusantara", Email: "smk@example.com"},
		plan: &schooldomain.Plan{
			ID:           planID,
			Name:         "Pro",
			MaxStudents:  1500,
			MonthlyPrice: 1299000,
			YearlyPrice:  12990000,
		},
		activeSubscription: &schooldomain.Subscription{
			Plan: &schooldomain.Plan{ID: uuid.New(), Name: "Basic", MaxStudents: 500},
		},
	}
	gateway := &fakeGateway{
		session: &GatewayCheckoutSession{
			Token:       "snap-token",
			RedirectURL: "https://midtrans.example/redirect",
			ExpiresAt:   timePtr(time.Now().Add(24 * time.Hour)),
		},
	}

	handler := NewCreateCheckoutHandler(repo, schools, &fakeStudentCounter{total: 400}, gateway)
	payment, err := handler.Handle(context.Background(), CreateCheckoutCommand{
		RequesterRole:     "admin_sekolah",
		RequesterSchoolID: &schoolID,
		RequesterUserID:   userID,
		SchoolID:          schoolID,
		PlanID:            planID,
		Cycle:             "month",
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if payment == nil || repo.created == nil {
		t.Fatal("expected pending payment to be created")
	}
	if payment.Status != billingdomain.PaymentStatusPending {
		t.Fatalf("expected pending status, got %s", payment.Status)
	}
	if gateway.receivedInput == nil || gateway.receivedInput.Amount != 1299000 {
		t.Fatalf("expected gateway amount 1299000, got %#v", gateway.receivedInput)
	}
}

func TestCreateCheckoutHandler_RejectsExistingPendingPayment(t *testing.T) {
	schoolID := uuid.New()
	handler := NewCreateCheckoutHandler(
		&fakePaymentRepo{
			findPending: &billingdomain.PaymentTransaction{
				ID:        uuid.New(),
				Status:    billingdomain.PaymentStatusPending,
				ExpiresAt: timePtr(time.Now().Add(1 * time.Hour)),
			},
		},
		&fakeBillingSchoolRepo{
			school: &schooldomain.School{ID: schoolID, Name: "SMK Nusantara"},
			plan:   &schooldomain.Plan{ID: uuid.New(), Name: "Pro", MaxStudents: 1500, MonthlyPrice: 1299000, YearlyPrice: 12990000},
			activeSubscription: &schooldomain.Subscription{
				Plan: &schooldomain.Plan{ID: uuid.New(), Name: "Basic", MaxStudents: 500},
			},
		},
		&fakeStudentCounter{total: 200},
		&fakeGateway{},
	)

	_, err := handler.Handle(context.Background(), CreateCheckoutCommand{
		RequesterRole:     "admin_sekolah",
		RequesterSchoolID: &schoolID,
		RequesterUserID:   uuid.New(),
		SchoolID:          schoolID,
		PlanID:            uuid.New(),
		Cycle:             "month",
	})
	if err == nil || !apperror.Is(err, apperror.ErrConflict) {
		t.Fatalf("expected pending conflict error, got %v", err)
	}
}

func TestHandleMidtransNotificationHandler_ActivatesSubscriptionOnPaidStatus(t *testing.T) {
	schoolID := uuid.New()
	planID := uuid.New()
	repo := &fakePaymentRepo{
		findByOrderID: &billingdomain.PaymentTransaction{
			ID:              uuid.New(),
			SchoolID:        schoolID,
			PlanID:          planID,
			Cycle:           "year",
			Amount:          12990000,
			Status:          billingdomain.PaymentStatusPending,
			ProviderOrderID: "EA-order",
		},
	}
	schools := &fakeBillingSchoolRepo{
		plan: &schooldomain.Plan{
			ID:           planID,
			Name:         "Pro",
			MaxStudents:  1500,
			MonthlyPrice: 1299000,
			YearlyPrice:  12990000,
		},
	}
	gateway := &fakeGateway{
		verifyResult: true,
		status: &GatewayTransactionStatus{
			OrderID:           "EA-order",
			TransactionID:     "trx-1",
			TransactionStatus: "settlement",
			StatusCode:        "200",
			GrossAmount:       "12990000.00",
			PaymentType:       "bank_transfer",
			RawResponse:       `{"transaction_status":"settlement"}`,
		},
	}

	handler := NewHandleMidtransNotificationHandler(repo, schools, &fakeStudentCounter{total: 300}, gateway)
	payment, err := handler.Handle(context.Background(), HandleMidtransNotificationCommand{
		OrderID:      "EA-order",
		StatusCode:   "200",
		GrossAmount:  "12990000.00",
		SignatureKey: "sig",
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if payment.Status != billingdomain.PaymentStatusPaid {
		t.Fatalf("expected paid status, got %s", payment.Status)
	}
	if payment.ActivatedSubscriptionID == nil {
		t.Fatal("expected activated subscription id to be set")
	}
	if schools.replacedSubscription == nil || schools.replacedSubscription.PlanID != planID {
		t.Fatalf("expected activated plan %s, got %#v", planID, schools.replacedSubscription)
	}
}

func TestHandleMidtransNotificationHandler_FallsBackToWebhookPayloadWhenStatusLookupFails(t *testing.T) {
	schoolID := uuid.New()
	planID := uuid.New()
	repo := &fakePaymentRepo{
		findByOrderID: &billingdomain.PaymentTransaction{
			ID:              uuid.New(),
			SchoolID:        schoolID,
			PlanID:          planID,
			Cycle:           "month",
			Amount:          499000,
			Status:          billingdomain.PaymentStatusPending,
			ProviderOrderID: "EA-order",
		},
	}
	schools := &fakeBillingSchoolRepo{
		plan: &schooldomain.Plan{
			ID:           planID,
			Name:         "Basic",
			MaxStudents:  500,
			MonthlyPrice: 499000,
			YearlyPrice:  4990000,
		},
	}
	gateway := &fakeGateway{
		verifyResult: true,
		statusErr:    apperror.New(apperror.ErrBadRequest, "failed to fetch Midtrans transaction status"),
	}

	handler := NewHandleMidtransNotificationHandler(repo, schools, &fakeStudentCounter{total: 100}, gateway)
	payment, err := handler.Handle(context.Background(), HandleMidtransNotificationCommand{
		OrderID:           "EA-order",
		StatusCode:        "200",
		GrossAmount:       "499000.00",
		SignatureKey:      "sig",
		TransactionID:     "trx-3",
		TransactionStatus: "settlement",
		PaymentType:       "qris",
		FraudStatus:       "accept",
		RawNotification:   `{"transaction_status":"settlement"}`,
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if payment.Status != billingdomain.PaymentStatusPaid {
		t.Fatalf("expected paid status, got %s", payment.Status)
	}
	if payment.ActivatedSubscriptionID == nil {
		t.Fatal("expected activated subscription id to be set")
	}
}

func TestGetPaymentHandler_RefreshesPendingMidtransPayment(t *testing.T) {
	schoolID := uuid.New()
	planID := uuid.New()
	paymentID := uuid.New()
	repo := &fakePaymentRepo{
		findByID: &billingdomain.PaymentTransaction{
			ID:              paymentID,
			SchoolID:        schoolID,
			PlanID:          planID,
			Cycle:           "month",
			Amount:          499000,
			Status:          billingdomain.PaymentStatusPending,
			Provider:        billingdomain.ProviderMidtrans,
			ProviderOrderID: "EA-order",
		},
	}
	schools := &fakeBillingSchoolRepo{
		plan: &schooldomain.Plan{
			ID:           planID,
			Name:         "Basic",
			MaxStudents:  500,
			MonthlyPrice: 499000,
			YearlyPrice:  4990000,
		},
	}
	gateway := &fakeGateway{
		status: &GatewayTransactionStatus{
			OrderID:           "EA-order",
			TransactionID:     "trx-2",
			TransactionStatus: "settlement",
			StatusCode:        "200",
			GrossAmount:       "499000.00",
			PaymentType:       "bank_transfer",
			RawResponse:       `{"transaction_status":"settlement"}`,
		},
	}

	handler := NewGetPaymentHandler(repo, schools, &fakeStudentCounter{total: 100}, gateway)
	payment, err := handler.Handle(context.Background(), GetPaymentQuery{
		RequesterRole: "superadmin",
		SchoolID:      schoolID,
		PaymentID:     paymentID,
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if payment.Status != billingdomain.PaymentStatusPaid {
		t.Fatalf("expected paid status, got %s", payment.Status)
	}
	if repo.updated == nil {
		t.Fatal("expected refreshed payment to be persisted")
	}
}

func timePtr(v time.Time) *time.Time { return &v }
