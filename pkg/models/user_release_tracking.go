package models

import "time"

type UserReleaseTracking struct {
	ID             string          `json:"id" bson:"_id,omitempty"`
	UserID         string          `json:"userId" bson:"userId"`
	ViewedReleases []ViewedRelease `json:"viewedReleases" bson:"viewedReleases"`
	LastChecked    time.Time       `json:"lastChecked" bson:"lastChecked"`
	Environment    string          `json:"environment" bson:"environment"`
}

type ViewedRelease struct {
	ReleaseID string    `json:"releaseId" bson:"releaseId"`
	ViewedAt  time.Time `json:"viewedAt" bson:"viewedAt"`
}

type GetUnviewedReleasesResponse struct {
	Releases   []Release `json:"releases"`
	TotalCount int       `json:"totalCount"`
}

type MarkViewedRequest struct {
	ReleaseIDs  []string `json:"releaseIds"`
	Environment string   `json:"environment"`
}

type GetUserStatusResponse struct {
	ViewedReleases []ViewedRelease `json:"viewedReleases"`
	LastChecked    time.Time       `json:"lastChecked"`
}
