package mailpit

import (
	"SwallowGo/internal/pgstore"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wneessen/go-mail"
)

type store interface {
	GetTrip(context.Context, uuid.UUID) (pgstore.Trip, error)
}

type Mailpit struct{
	store store
}

func NewMailpit(pool *pgxpool.Pool) Mailpit {
	return Mailpit{pgstore.New(pool)}
}

func (mp Mailpit) SendConfirmTripEmailToTripOwner(tripID uuid.UUID) error {
	ctx := context.Background()
	trip, err := mp.store.GetTrip(ctx, tripID)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for SendConfirmTripEmailToTripOwner: %w", err)
	}

	title, body, err := textMail(
		"Confirm your trip",
		fmt.Sprintf(`
			Hello, %s!
	
			Your trip to %s, which starts on %s, needs to be confirmed.
			Click the button below to confirm.
		`,
			trip.OwnerName, trip.Destination, trip.StartsAt.Time.Format(time.DateOnly),
		),
	)
	if err != nil {
		return fmt.Errorf("mailpit: failed create email body SendConfirmTripEmailToTripOwner: %w", err)
	}

	if err := sendMail(title,body,trip.OwnerEmail); err != nil {
		return fmt.Errorf("mailpit: failed create email body SendConfirmTripEmailToTripOwner: %w", err)
	}
	return nil
}

func (mp Mailpit) ReSendConfirmTripEmailToTripOwner(tripID uuid.UUID) error {
	ctx := context.Background()
	trip, err := mp.store.GetTrip(ctx, tripID)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for ReSendConfirmTripEmailToTripOwner: %w", err)
	}

	title, body, err := textMail(
		"Reconfirm your trip",
		fmt.Sprintf(`
			Hello, %s!

			Your trip to %s, which starts on %s, requires reconfirmation.
			Please click the button below to reconfirm your trip.
		`,
			trip.OwnerName, trip.Destination, trip.StartsAt.Time.Format(time.DateOnly),
		),
	)
	if err != nil {
		return fmt.Errorf("mailpit: failed create email body ReSendConfirmTripEmailToTripOwner: %w", err)
	}

	if err := sendMail(title,body,trip.OwnerEmail); err != nil {
		return fmt.Errorf("mailpit: failed create email body ReSendConfirmTripEmailToTripOwner: %w", err)
	}
	return nil
}

func sendMail(title string, body string, email string) error{
	msg := mail.NewMsg()
	if err := msg.From("contact@swallowgo.com"); err != nil {
		return err
	}

	if err := msg.To(email); err != nil {
		return err
	}

	msg.Subject(title)
	msg.SetBodyString(mail.TypeTextPlain, body)

	client, err := mail.NewClient(os.Getenv("SWALLOWGO_EMAIL_HOST"), mail.WithTLSPortPolicy(mail.NoTLS), mail.WithPort(1025))
	if err != nil {
		return err
	}

	if err := client.DialAndSend(msg); err != nil {
		return err
	}

	return nil
}

func textMail(title string, body string) (string, string, error) {
    return title, body, nil
}