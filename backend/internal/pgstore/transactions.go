package pgstore

import (
	"SwallowGo/internal/api/spec"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (q *Queries) CreateTrip(ctx context.Context, pool *pgxpool.Pool, params spec.CreateTripRequest) (uuid.UUID, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to begin tx for CreateTrip: %w", err)
	}

	defer func() { _ = tx.Rollback(ctx) }()

	qtx := q.WithTx(tx)
	tripID, err := qtx.InsertTrip(ctx, InsertTripParams{
		Destination: params.Destination,
		OwnerEmail:  string(params.OwnerEmail),
		OwnerName:   params.OwnerName,
		StartsAt:    pgtype.Timestamp{Valid: true, Time: params.StartsAt},
		EndsAt:      pgtype.Timestamp{Valid: true, Time: params.EndsAt},
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to insert trip for CreateTrip: %w", err)
	}

	participants := make([]InviteParticipantsToTripParams, len(params.EmailsToInvite))
	for i, eti := range params.EmailsToInvite {
		participants[i] = InviteParticipantsToTripParams{
			TripID: tripID,
			Email:  string(eti),
		}
	}

	if _, err := qtx.InviteParticipantsToTrip(ctx, participants); err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to insert participants for CreateTrip: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to insert trip for CreateTrip: %w", err)
	}

	return tripID, nil
}

func (q *Queries) PutTrip(ctx context.Context, pool *pgxpool.Pool, params spec.UpdateTripRequest, tripID uuid.UUID) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("pgstore: failed to begin tx for PutTrip: %w", err)
	}

	defer func() { _ = tx.Rollback(ctx) }()

	qtx := q.WithTx(tx)
	if err := qtx.UpdateTrip(ctx, UpdateTripParams{
		Destination: params.Destination,
		StartsAt:    pgtype.Timestamp{Valid: true, Time: params.StartsAt},
		EndsAt:      pgtype.Timestamp{Valid: true, Time: params.EndsAt},
		ID: tripID,
	}); err != nil {
		return fmt.Errorf("pgstore: failed to update trip for UpdateTrip: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("pgstore: failed to commit transaction for UpdateTrip: %w", err)
	}

	return nil
}

func (q *Queries) InsertActivity(ctx context.Context, pool *pgxpool.Pool, params spec.CreateActivityRequest, tripID uuid.UUID) (uuid.UUID, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to begin tx for InsertActivity: %w", err)
	}

	defer func() { _ = tx.Rollback(ctx) }()

	qtx := q.WithTx(tx)
	activityID, err := qtx.CreateActivity(ctx, CreateActivityParams{
		TripID: tripID,
		Title: params.Title,
		OccursAt: pgtype.Timestamp{Valid: true, Time: params.OccursAt},
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to insert activity for CreateActivity: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, fmt.Errorf("pgstore: failed to insert trip for CreateTrip: %w", err)
	}

	return activityID, nil
}