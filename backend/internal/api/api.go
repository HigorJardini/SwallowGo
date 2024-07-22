package api

import (
	"SwallowGo/internal/api/spec"
	"SwallowGo/internal/pgstore"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/discord-gophers/goapi-gen/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.uber.org/zap"
)

type store interface {
	// trip
	GetTrip(ctx context.Context, tripID uuid.UUID) (pgstore.Trip, error)
	CreateTrip(context.Context, *pgxpool.Pool, spec.CreateTripRequest) (uuid.UUID, error)
	PutTrip(ctx context.Context, pool *pgxpool.Pool, params spec.UpdateTripRequest, tripID uuid.UUID) error
	ConfirmTrip(ctx context.Context, tripID uuid.UUID) error
	//Participant
	GetParticipant(ctx context.Context, participantID uuid.UUID) (pgstore.Participant, error)
	ConfirmParticipant(ctx context.Context, participantID uuid.UUID) error
	GetParticipants(ctx context.Context, tripID uuid.UUID) ([]pgstore.Participant, error)
	InsertInviteParticipantToTrip(ctx context.Context, pool *pgxpool.Pool, params spec.InviteParticipantRequest, tripID uuid.UUID) (uuid.UUID, error)
	//Activities
    InsertActivity(ctx context.Context, pool *pgxpool.Pool, params spec.CreateActivityRequest, tripID uuid.UUID) (uuid.UUID, error)
	GetTripActivities(ctx context.Context, tripID uuid.UUID) ([]pgstore.Activity, error)
	//Links
	GetTripLinks(ctx context.Context, tripID uuid.UUID) ([]pgstore.Link, error)
	InsertTripsTripIDLinks(ctx context.Context, pool *pgxpool.Pool, params spec.CreateLinkRequest, tripID uuid.UUID) (uuid.UUID, error)
}

type mailer interface {
	SendConfirmTripEmailToTripOwner(tripID uuid.UUID) error
	ReSendConfirmTripEmailToTripOwner(tripID uuid.UUID) error
}

type API struct {
	store store
	logger *zap.Logger
	validator *validator.Validate
	pool *pgxpool.Pool
	mailer    mailer
}

func NewApi(pool *pgxpool.Pool, logger *zap.Logger, mailer mailer) API {
	validator := validator.New(validator.WithRequiredStructEnabled())
	return API{pgstore.New(pool), logger, validator, pool, mailer}
}

// Confirms a participant on a trip.
// (PATCH /participants/{participantId}/confirm)
func (api API) PatchParticipantsParticipantIDConfirm(w http.ResponseWriter, r *http.Request, participantID string) *spec.Response {
	id, err := uuid.Parse(participantID)
	if err != nil {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "invalid uuid",})
	}

	participant, err := api.store.GetParticipant(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "participant not found",})
		}
		api.logger.Error("failed to get participant", zap.Error(err), zap.String("participant_id", participantID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "something went wrong, try again",})
	}

	if participant.IsConfirmed {
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "participant already confirmed",})
	}

	if err := api.store.ConfirmParticipant(r.Context(), id); err != nil {
		api.logger.Error("failed to confim participant", zap.Error(err), zap.String("participant_id", participantID))
		return spec.PatchParticipantsParticipantIDConfirmJSON400Response(spec.Error{Message: "something went wrong, try again",})
	}

	return spec.PatchParticipantsParticipantIDConfirmJSON204Response(nil)
}

// Create a new trip
// (POST /trips)
func (api API) PostTrips(w http.ResponseWriter, r *http.Request) *spec.Response {
	var body spec.CreateTripRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		spec.PostTripsJSON400Response(spec.Error{Message: "invalid json: " + err.Error()})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "invalid input: " + err.Error()})
	}

	tripID, err := api.store.CreateTrip(r.Context(), api.pool, body)
	if err != nil {
		return spec.PostTripsJSON400Response(spec.Error{Message: "Failed to create trip, try again"})
	}

	go func() {
		if err := api.mailer.SendConfirmTripEmailToTripOwner(tripID); err != nil {
			api.logger.Error("failed to send email on PostTrips", zap.Error(err), zap.String("trip_id", tripID.String()))
		}
	} ()

	return spec.PostTripsJSON201Response(spec.CreateTripResponse{TripID: tripID.String()});
}

// Get a trip details.
// (GET /trips/{tripId})
func (api API) GetTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "invalid uuid",})
	}

	trip, err := api.store.GetTrip(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "trip not found",})
		}
		api.logger.Error("failed to get trip", zap.Error(err), zap.String("trip_id", tripID))
		return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "something went wrong, try again",})
	}

	return spec.GetTripsTripIDJSON200Response(spec.GetTripDetailsResponse{Trip: spec.GetTripDetailsResponseTripObj{
		Destination: trip.Destination,
		EndsAt:      trip.EndsAt.Time,
		ID:          trip.ID.String(),
		IsConfirmed: trip.IsConfirmed,
		StartsAt:    trip.StartsAt.Time,
	}});
}

// Update a trip.
// (PUT /trips/{tripId})
func (api API) PutTripsTripID(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	var body spec.UpdateTripRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		spec.PutTripsTripIDJSON400Response(spec.Error{Message: "invalid json: " + err.Error()})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "invalid input: " + err.Error()})
	}

	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "invalid uuid",})
	}

	if _, err := api.store.GetTrip(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "trip not found",})
		}
		api.logger.Error("failed to get trip", zap.Error(err), zap.String("trip_id", tripID))
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "something went wrong, try again",})
	}

	if err := api.store.PutTrip(r.Context(), api.pool, body, id); err != nil {
		api.logger.Error("failed to update trip", zap.Error(err), zap.String("trip_id", tripID))
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "something went wrong, try again",})
	}

	go func() {
		if err := api.mailer.ReSendConfirmTripEmailToTripOwner(id); err != nil {
			api.logger.Error("failed to send email on PutTripsTripID", zap.Error(err), zap.String("trip_id", id.String()))
		}
	} ()

	return spec.PutTripsTripIDJSON204Response(nil);
}

// Get a trip activities.
// (GET /trips/{tripId}/activities)
func (api API) GetTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDActivitiesJSON400Response(spec.Error{Message: "invalid uuid",})
	}

	activities, err := api.store.GetTripActivities(r.Context(), id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "activities not found"})
        }
        api.logger.Error("failed to get activities", zap.Error(err), zap.String("trip_id", tripID))
        return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "something went wrong, try again"})
    }

	// Group activities by date
    activityMap := make(map[time.Time][]spec.GetTripActivitiesResponseInnerArray)
    for _, activity := range activities {
        if !activity.OccursAt.Valid {
            api.logger.Error("invalid timestamp", zap.String("activity_id", activity.ID.String()))
            return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "failed to process activities"})
        }

        occursAt := activity.OccursAt.Time
        date := time.Date(occursAt.Year(), occursAt.Month(), occursAt.Day(), 0, 0, 0, 0, occursAt.Location())
        innerActivity := spec.GetTripActivitiesResponseInnerArray{
            ID:       activity.ID.String(),
            Title:    activity.Title,
            OccursAt: occursAt,
        }
        activityMap[date] = append(activityMap[date], innerActivity)
    }

	var arrActivities spec.GetTripActivitiesResponse
    for date, innerActivities := range activityMap {
        outerActivity := spec.GetTripActivitiesResponseOuterArray{
            Date:       date,
            Activities: innerActivities,
        }
        arrActivities.Activities = append(arrActivities.Activities, outerActivity)
    }

	return spec.GetTripsTripIDActivitiesJSON200Response(arrActivities)
}

// Create a trip activity.
// (POST /trips/{tripId}/activities)
func (api API) PostTripsTripIDActivities(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	var body spec.CreateActivityRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "invalid json: " + err.Error()})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "invalid input: " + err.Error()})
	}

	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PutTripsTripIDJSON400Response(spec.Error{Message: "invalid uuid",})
	}

	activityID, err := api.store.InsertActivity(r.Context(), api.pool, body, id)
	if err != nil {
		return spec.PostTripsTripIDActivitiesJSON400Response(spec.Error{Message: "Failed to create activity, try again"})
	}

	return spec.PostTripsTripIDActivitiesJSON201Response(spec.CreateActivityResponse{ActivityID: activityID.String()});
}

// Confirm a trip and send e-mail invitations.
// (GET /trips/{tripId}/confirm)
func (api API) GetTripsTripIDConfirm(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "invalid uuid",})
	}

	trip, err := api.store.GetTrip(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "trip not found",})
		}
		api.logger.Error("failed to get trip", zap.Error(err), zap.String("trip_id", tripID))
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "something went wrong, try again",})
	}

	if trip.IsConfirmed {
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "trip already confirmed",})
	}

	if err := api.store.ConfirmTrip(r.Context(), id); err != nil {
		api.logger.Error("failed to confim trip", zap.Error(err), zap.String("trip_id", tripID))
		return spec.GetTripsTripIDConfirmJSON400Response(spec.Error{Message: "something went wrong, try again",})
	}

	// Send email

	return spec.GetTripsTripIDConfirmJSON204Response(nil)
}

// Invite someone to the trip.
// (POST /trips/{tripId}/invites)
func (api API) PostTripsTripIDInvites(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	var body spec.InviteParticipantRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "invalid json: " + err.Error()})
	}

	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "invalid uuid",})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "invalid input: " + err.Error()})
	}

	participantID, err := api.store.InsertInviteParticipantToTrip(r.Context(), api.pool, body, id)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "Failed to insert invite participant, try again"})
	}

	// Send email
	
	return spec.PostTripsTripIDInvitesJSON201Response(spec.InviteParticipantResponse{ParticipantID: participantID.String()});
}

// Get a trip links.
// (GET /trips/{tripId}/links)
func (api API) GetTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "invalid uuid",})
	}

	links, err := api.store.GetTripLinks(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "participants not found",})
		}
		api.logger.Error("failed to get participants", zap.Error(err), zap.String("trip_id", tripID))
		return spec.PostTripsTripIDInvitesJSON400Response(spec.Error{Message: "something went wrong, try again",})
	}

	arrLinks := make([]spec.GetLinksResponseArray, len(links))
	for i, link := range links {
		arrLinks[i] = spec.GetLinksResponseArray{
			ID:          link.ID.String(),
			Title:       link.Title,
			URL:         link.Url,
		}
	}

	return spec.GetTripsTripIDLinksJSON200Response(spec.GetLinksResponse{
		Links: arrLinks,
	})
}

// Create a trip link.
// (POST /trips/{tripId}/links)
func (api API) PostTripsTripIDLinks(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	var body spec.CreateLinkRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "invalid json: " + err.Error()})
	}

	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "invalid uuid",})
	}

	if err := api.validator.Struct(body); err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "invalid input: " + err.Error()})
	}

	linkId, err := api.store.InsertTripsTripIDLinks(r.Context(), api.pool, body, id)
	if err != nil {
		return spec.PostTripsTripIDLinksJSON400Response(spec.Error{Message: "Failed to insert trip link, try again"})
	}
	
	return spec.PostTripsTripIDLinksJSON201Response(spec.CreateLinkResponse{LinkID: linkId.String()});
}

// Get a trip participants.
// (GET /trips/{tripId}/participants)
func (api API) GetTripsTripIDParticipants(w http.ResponseWriter, r *http.Request, tripID string) *spec.Response {
	id, err := uuid.Parse(tripID)
	if err != nil {
		return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "invalid uuid",})
	}

	participants, err := api.store.GetParticipants(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return spec.GetTripsTripIDParticipantsJSON400Response(spec.Error{Message: "participants not found",})
		}
		api.logger.Error("failed to get participants", zap.Error(err), zap.String("trip_id", tripID))
		return spec.GetTripsTripIDJSON400Response(spec.Error{Message: "something went wrong, try again",})
	}
	

	arrParticipants := make([]spec.GetTripParticipantsResponseArray, len(participants))
	for i, participant := range participants {
		arrParticipants[i] = spec.GetTripParticipantsResponseArray{
			Email:       types.Email(participant.Email),
			ID:          participant.ID.String(),
			IsConfirmed: participant.IsConfirmed,
			Name:        nil,
		}
	}

	return spec.GetTripsTripIDParticipantsJSON200Response(spec.GetTripParticipantsResponse{
		Participants: arrParticipants,
	})
}

