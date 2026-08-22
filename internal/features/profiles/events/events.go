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
	ProfileAvatarUpdated       eventType = "profile_avatar_updated_event"
	ProfileBioUpdated          eventType = "profile_bio_updated_event"
	ProfileImageUploaded       eventType = "profile_image_uploaded_event"
	ProfileHeaderImageUploaded eventType = "profile_header_image_uploaded_event"
)

const (
	FieldProfileAvatarUpdatedEventID       = "profile_avatar_updated_event_id"
	FieldProfileBioUpdatedEventID          = "profile_bio_updated_event_id"
	FieldProfileImageUploadedEventID       = "profile_image_uploaded_event_id"
	FieldProfileHeaderImageUploadedEventID = "profile_header_image_uploaded_event_id"
)

// profile event fields
const (
	FieldProfileAvatarUpdatedAvatar    = "avatar"
	FieldProfileAvatarUpdatedUpdatedAt = "updated_at"
	FieldProfileBioUpdatedBio          = "bio"
	FieldProfileBioUpdatedBioHash      = "bio_hash"
	FieldProfileBioUpdatedAt           = "updated_at"
	FieldProfileImageURL               = "image_url"
	FieldProfileImageUploadedAt        = "uploaded_at"
	FieldScopeUserRegisteredEventID    = "scope.user_registered_event_id"
)

type ProfileAvatarUpdatedEvent struct {
	EventID   string           `json:"profile_avatar_updated_event_id"`
	Avatar    string           `json:"avatar"`
	UpdatedAt string           `json:"updated_at"`
	Scope     ProfileUserScope `json:"scope"`
}

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
	UserRegisteredEventID string `json:"user_registered_event_id"`
}

func NewProfileBioUpdatedEvent(
	eventID,
	bio string,
	updatedAt time.Time,
	userRegisteredEventID string,
	subjectKey protectedpii.SubjectDataKey,
	metadata map[string]any,
) eventstore.DomainEvent {
	protector := protectedpii.FromEnv()
	event := ProfileBioUpdatedEvent{
		EventID:   eventID,
		Bio:       protector.MustProtectWithDataKey(bio, FieldProfileBioUpdatedBio, subjectKey),
		BioHash:   protector.BlindIndex(FieldProfileBioUpdatedBio, bio),
		UpdatedAt: updatedAt.Format(time.RFC3339),
		Scope:     ProfileUserScope{UserRegisteredEventID: userRegisteredEventID},
	}
	return eventstore.DomainEvent{
		EventID:   eventID,
		EventType: ProfileBioUpdated,
		Data:      eventstore.MustData(event),
		Metadata:  metadata,
	}
}

func NewProfileAvatarUpdatedEvent(
	eventID,
	avatar string,
	updatedAt time.Time,
	userRegisteredEventID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ProfileAvatarUpdatedEvent{
		EventID:   eventID,
		Avatar:    avatar,
		UpdatedAt: updatedAt.Format(time.RFC3339),
		Scope:     ProfileUserScope{UserRegisteredEventID: userRegisteredEventID},
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
	userRegisteredEventID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ProfileImageUploadedEvent{
		EventID:    eventID,
		ImageURL:   imageURL,
		UploadedAt: uploadedAt.Format(time.RFC3339),
		Scope:      ProfileUserScope{UserRegisteredEventID: userRegisteredEventID},
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
	userRegisteredEventID string,
	metadata map[string]any,
) eventstore.DomainEvent {
	event := ProfileHeaderImageUploadedEvent{
		EventID:    eventID,
		ImageURL:   imageURL,
		UploadedAt: uploadedAt.Format(time.RFC3339),
		Scope:      ProfileUserScope{UserRegisteredEventID: userRegisteredEventID},
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
