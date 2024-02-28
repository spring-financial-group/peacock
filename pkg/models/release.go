package models

import "time"

type Release struct {
	CreatedAt    time.Time     `json:"createdAt" bson:"createdAt"`
	ReleaseNotes []ReleaseNote `json:"releaseNotes" bson:"releaseNotes"`
	Environment  string        `json:"environment" bson:"environment"`
}

type GetReleasesResponse struct {
	Releases []Release `json:"releases"`
}
