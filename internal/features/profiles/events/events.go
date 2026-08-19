package events

import (
	"time"

	"seek/internal/eventstore"
	"seek/internal/protectedpii"
)

type eventType = eventstore.EventType

var eventTypeKey = eventstore.EventTypeKey

// profile event types
const (
	ProfileBioUpdated          = "ProfileBioUpdated"
	ProfileImageUploaded       = "ProfileImageUploaded"
	ProfileHeaderImageUploaded = "ProfileHeaderImageUploaded"
)

// profile event fields
const (
	ProfileBioUpdatedEventIDField          = "profile_bio_updated_event_id"
	ProfileBioUpdatedBioField              = "bio"
	ProfileBioUpdatedBioHashField          = "bio_hash"
	ProfileBioUpdatedAtField               = "updated_at"
	ProfileImageUploadedEventIDField       = "profile_image_uploaded_event_id"
	ProfileHeaderImageUploadedEventIDField = "profile_header_image_uploaded_event_id"
	ProfileImageURLField                   = "image_url"
	ProfileImageUploadedAtField            = "uploaded_at"
	ProfileScopeUserRegisteredIDField      = "scope.user_registered_id"
)

type ProfileBioUpdatedEvent struct {
	EventID   string             `json:"profile_bio_updated_event_id"`
	Bio       protectedpii.Value `json:"bio"`
	BioHash   string             `json:"bio_hash"`
	UpdatedAt string             `json:"updated_at"`
	Scope     ProfileUserScope   `json:"scope"`
}

type ProfileImageUploadedEvent struct {
	EventID    string           `json:"profile_image_uploaded_event_id"`
	ImageURL   string           `json:"image_url"`
	UploadedAt string           `json:"uploaded_at"`
	Scope      ProfileUserScope `json:"scope"`
}

type ProfileHeaderImageUploadedEvent struct {
	EventID    string           `json:"profile_header_image_uploaded_event_id"`
	ImageURL   string           `json:"image_url"`
	UploadedAt string           `json:"uploaded_at"`
	Scope      ProfileUserScope `json:"scope"`
}

type ProfileUserScope struct {
	UserRegisteredID string `json:"user_registered_id"`
}

func NewProfileBioUpdatedEvent(
	eventID,
	bio string,
	updatedAt time.Time,
	userRegisteredID string,
	subjectKey protectedpii.SubjectDataKey,
	metadata map[string]any,
) eventstore.DomainEvent {
	protector := protectedpii.FromEnv()
	event := ProfileBioUpdatedEvent{
		EventID:   eventID,
		Bio:       protector.MustProtectWithDataKey(bio, ProfileBioUpdatedBioField, subjectKey),
		BioHash:   protector.BlindIndex(ProfileBioUpdatedBioField, bio),
		UpdatedAt: updatedAt.Format(time.RFC3339),
		Scope:     ProfileUserScope{UserRegisteredID: userRegisteredID},
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: ProfileBioUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewProfileImageUploadedEvent(
	eventID,
	imageURL string,
	uploadedAt time.Time,
	userRegisteredID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ProfileImageUploadedEvent{
		EventID:    eventID,
		ImageURL:   imageURL,
		UploadedAt: uploadedAt.Format(time.RFC3339),
		Scope:      ProfileUserScope{UserRegisteredID: userRegisteredID},
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: ProfileImageUploaded,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewProfileHeaderImageUploadedEvent(
	eventID,
	imageURL string,
	uploadedAt time.Time,
	userRegisteredID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ProfileHeaderImageUploadedEvent{
		EventID:    eventID,
		ImageURL:   imageURL,
		UploadedAt: uploadedAt.Format(time.RFC3339),
		Scope:      ProfileUserScope{UserRegisteredID: userRegisteredID},
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: ProfileHeaderImageUploaded,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func Channel(userRegisteredID string) string {
	return "profiles." + userRegisteredID
}
