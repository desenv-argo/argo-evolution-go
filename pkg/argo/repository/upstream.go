package argo_repository

import (
	"context"
	"encoding/json"

	argo_model "github.com/evolution-foundation/evolution-go/pkg/argo/model"
)

func (r *repository) SaveUpstreamSnapshot(ctx context.Context, snapshot *argo_model.UpstreamSnapshot) error {
	changes, err := json.Marshal(snapshot.Changes)
	if err != nil {
		return err
	}
	snapshot.ChangesJSON = string(changes)
	return r.db.WithContext(ctx).Create(snapshot).Error
}

func (r *repository) LatestUpstreamSnapshot(ctx context.Context) (*argo_model.UpstreamSnapshot, error) {
	var snapshot argo_model.UpstreamSnapshot
	err := r.db.WithContext(ctx).Order("checked_at DESC, created_at DESC").Limit(1).Find(&snapshot).Error
	if err != nil {
		return nil, err
	}
	if snapshot.ID == "" {
		return nil, nil
	}
	if snapshot.ChangesJSON != "" {
		if err := json.Unmarshal([]byte(snapshot.ChangesJSON), &snapshot.Changes); err != nil {
			return nil, err
		}
	}
	if snapshot.Changes == nil {
		snapshot.Changes = []argo_model.UpstreamChange{}
	}
	return &snapshot, nil
}
