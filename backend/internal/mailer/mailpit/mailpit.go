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
	//Trips
	GetTrip(context.Context, uuid.UUID) (pgstore.Trip, error)
	//Participants
	GetParticipant(ctx context.Context, participantID uuid.UUID) (pgstore.Participant, error)
	GetParticipants(ctx context.Context, tripID uuid.UUID) ([]pgstore.Participant, error)
	
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
		fmt.Sprintf(`SwallowGo - Action Needed: Confirm Your Trip To %s!`,trip.Destination,),
		fmt.Sprintf(`
			Hello, %s!
	
			Your trip to %s, starting on %s, needs to be confirmed.

			Please click the button below to confirm your travel plans.

			Thank you for your prompt attention,
			SwallowGo
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
		fmt.Sprintf(`SwallowGo - Action Required: Reconfirm Your Trip To %s!`,trip.Destination,),
		fmt.Sprintf(`
			Hello, %s!

			We need your attention regarding your upcoming trip to %s, starting on %s. Due to recent changes, we require you to reconfirm your travel plans.

			Please click the button below to reconfirm your trip.

			Thank you for your cooperation,
			SwallowGo
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

func (mp Mailpit) SendConfirmTripEmailToTripinvitations(tripID uuid.UUID) error {
	ctx := context.Background()
	participants, err := mp.store.GetParticipants(ctx, tripID)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for SendConfirmTripEmailToTripinvitations: %w", err)
	}
	trip, err := mp.store.GetTrip(ctx, tripID)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for SendConfirmTripEmailToTripOwner: %w", err)
	}

	for _, participant := range participants {
		title, body, err := textMail(
			fmt.Sprintf(`SwallowGo - Your Trip to %s is Confirmed!`,trip.Destination,),
			fmt.Sprintf(`
				Hello, %s!

				We are excited to confirm your upcoming trip to %s, starting on %s. We hope you have a fantastic journey filled with unforgettable experiences.

				Safe travels,
				SwallowGo
			`,
				participant.Email, trip.Destination, trip.StartsAt.Time.Format(time.DateOnly),
			),
		)
		if err != nil {
			return fmt.Errorf("mailpit: failed create email body SendConfirmTripEmailToTripinvitations: %w", err)
		}

		if err := sendMail(title,body,trip.OwnerEmail); err != nil {
			return fmt.Errorf("mailpit: failed create email body SendConfirmTripEmailToTripinvitations: %w", err)
		}
	}

	return nil
}

func (mp Mailpit) SendConfirmTripEmailToTripinvitation(tripID uuid.UUID,ParticipantID uuid.UUID) error {
	ctx := context.Background()
	participant, err := mp.store.GetParticipant(ctx, ParticipantID)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for SendConfirmTripEmailToTripinvitations: %w", err)
	}
	trip, err := mp.store.GetTrip(ctx, tripID)
	if err != nil {
		return fmt.Errorf("mailpit: failed to get trip for SendConfirmTripEmailToTripOwner: %w", err)
	}

	title, body, err := textMail(
		fmt.Sprintf(`SwallowGo - Your Trip to %s is Confirmed!`,trip.Destination,),
		fmt.Sprintf(`
			Hello, %s!

			We are excited to confirm your upcoming trip to %s, starting on %s. We hope you have a fantastic journey filled with unforgettable experiences.

			Safe travels,
			SwallowGo
		`,
			participant.Email, trip.Destination, trip.StartsAt.Time.Format(time.DateOnly),
		),
	)
	
	if err != nil {
		return fmt.Errorf("mailpit: failed create email body SendConfirmTripEmailToTripinvitations: %w", err)
	}

	if err := sendMail(title,body,trip.OwnerEmail); err != nil {
		return fmt.Errorf("mailpit: failed create email body SendConfirmTripEmailToTripinvitations: %w", err)
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