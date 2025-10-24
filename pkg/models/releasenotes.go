package models

import "fmt"

type ReleaseNote struct {
	Teams   Teams  `json:"teams"`
	Content string `json:"content"`
}

func (r *ReleaseNote) AppendContent(content string) {
	r.Content += fmt.Sprintf("\n\n---\n\n%s", content)
}

func (r *ReleaseNote) AreTeamsEqual(other ReleaseNote) bool {
	if len(r.Teams) != len(other.Teams) {
		return false
	}
	for i, team := range r.Teams {
		if team.Name != other.Teams[i].Name {
			return false
		}
	}
	return true
}
