package tasks

import (
	"context"
	"mashaghel/internal/database/scylla"
	"testing"
	"time"

	"go.uber.org/zap"
)

func Test_task_processWatched(t *testing.T) {
	type fields struct {
		scylla scylla.ScyllaDB
		logger *zap.Logger
	}
	type args struct {
		ctx context.Context
	}
	testCases := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "verify data consistency between recent_watch and ordered_watch tables",
			fields: fields{
				scylla: scyllaDB,
				logger: zap.NewExample(),
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			recentQuery := "SELECT profile_id, play_id, duration, watched_at FROM recent_watch"
			orderedQquery := "SELECT profile_id, play_id, duration, watched_at FROM ordered_watch"
			recentRows, err := tt.fields.scylla.Session().Query(recentQuery).WithContext(tt.args.ctx).Iter().SliceMap()
			if err != nil {
				t.Errorf("failed to execute query: %v", err)
				return
			}
			taskInstance := createTaskInstance()
			taskInstance.Start()
			// wait for tasks
			time.Sleep(10 * time.Second)

			orderedRows, err := tt.fields.scylla.Session().Query(orderedQquery).WithContext(tt.args.ctx).Iter().SliceMap()
			if err != nil {
				t.Errorf("failed to execute query: %v", err)
				return
			}

			profileIDSet := make(map[interface{}]struct{})
			for _, value := range recentRows {
				profileIDSet[value["profile_id"]] = struct{}{}
			}

			for _, row := range orderedRows {
				if _, exists := profileIDSet[row["profile_id"]]; exists {
					continue
				} else {
					t.Errorf("No match for profile_id: %v", row["profile_id"])
					return
				}
			}

			recentRows, err = tt.fields.scylla.Session().Query(recentQuery).WithContext(tt.args.ctx).Iter().SliceMap()
			if err != nil {
				t.Errorf("failed to execute query: %v", err)
				return
			}

			if len(recentRows) != 0 {
				t.Errorf("failed to delete recent watches: %v", err)
			}

		})
	}
}
