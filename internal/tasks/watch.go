package tasks

import (
	"fmt"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type playInfo struct {
	watchedAt  gocql.UUID
	duration   int
	profile_id string
	play_id    string
}

const (
	queryUpdateWatched              = `UPDATE watched SET watched_at = ? WHERE profile_id = ? AND play_id = ?;`
	queryInsertOrderedWatch         = `INSERT INTO ordered_watch (profile_id, play_id, duration, watched_at) VALUES (?, ?, ?, ?);`
	queryDeleteOutdatedOrderedWatch = `DELETE FROM ordered_watch WHERE profile_id = ? AND watched_at = ?;`
	queryDeleteOutdatedRecentWatch  = `DELETE FROM recent_watch USING TIMESTAMP ? WHERE profile_id = ? AND play_id = ?;`
)

func (t *task) processWatched() {
	profileToken := int64(0)
	var err error
	var playInfos map[string]playInfo
	wg := sync.WaitGroup{}
	daysAgo := time.Now().Add(time.Duration(t.configs.TasksConfig.WatchAgeLimit) * time.Hour)

	t.logger.Info("Starting processWatched task Powered by (YA ALI)")

	for {
		t.logger.Info("Processing idle watches", zap.Int64("profileToken", profileToken))
		playInfos, profileToken, err = t.processIdleWatches(profileToken, daysAgo)
		if err != nil {
			t.logger.Error("Error processing idle watches", zap.Error(err))
			time.Sleep(time.Second * 10)
			continue
		}
		if playInfos == nil {
			t.logger.Info("No idle watches found, job is done")
			wg.Wait()
			return
		}
		t.logger.Info("Idle watches found", zap.Int("count", len(playInfos)))
		wg.Add(1)
		t.workerpool.Submit(func() {
			t.updateWatches(playInfos, profileToken)
			wg.Done()
		})
	}
}

func (t *task) updateWatches(playInfos map[string]playInfo, profileToken int64) {
	t.logger.Info("Starting updateWatches", zap.Int64("profileToken", profileToken), zap.Int("playInfosCount", len(playInfos)))

	batch := t.scylla.Session().NewBatch(gocql.LoggedBatch)
	deleteTS := time.Now().UnixNano() / 1000 // this is for deleting recent_watches
	batch.SetConsistency(gocql.Any)

	placeHolder := ""
	i := 0

	for playId := range playInfos {
		if i == 0 {
			placeHolder += playId
		} else {
			placeHolder += "," + playId
		}
		i++
	}

	if i == 0 {
		t.logger.Warn("No playInfos to process for the given token", zap.Int64("profileToken", profileToken))
		return
	}

	t.logger.Info("Querying watched table", zap.String("placeHolder", placeHolder))

	var watchedAt gocql.UUID
	var play_id string

	query := fmt.Sprintf(`SELECT play_id, watched_at FROM watched WHERE token(profile_id) = ? AND play_id in (%v)`, placeHolder)
	iter := t.scylla.Session().Query(query, profileToken).Consistency(gocql.One).Iter()
	for iter.Scan(&play_id, &watchedAt) {
		playInfo, ok := playInfos[play_id]

		if !ok {
			t.logger.Error("Play info not found", zap.String("play_id", play_id))
			continue
		}

		t.logger.Info("Processing play info", zap.String("play_id", play_id), zap.String("profile_id", playInfo.profile_id))

		batch.Query(
			queryUpdateWatched,
			playInfo.watchedAt,
			playInfo.profile_id,
			playInfo.play_id,
		)
		batch.Query(
			queryInsertOrderedWatch,
			playInfo.profile_id,
			playInfo.play_id,
			playInfo.duration,
			playInfo.watchedAt,
		)
		if watchedAt != gocql.UUID(uuid.Nil) {
			batch.Query(queryDeleteOutdatedOrderedWatch, playInfo.profile_id, watchedAt)
		}
		batch.Query(queryDeleteOutdatedRecentWatch, deleteTS, playInfo.profile_id, playInfo.play_id)
	}

	if err := t.scylla.Session().ExecuteBatch(batch); err != nil {
		t.logger.Error("Error executing batch", zap.Error(err), zap.String("placeHolder", placeHolder))
		return
	}

	if err := iter.Close(); err != nil {
		t.logger.Error("Error closing iterator", zap.Error(err))
		return
	}

	t.logger.Info("Successfully updated watches")
}

func (t *task) processIdleWatches(token int64, daysAgo time.Time) (map[string]playInfo, int64, error) {
	t.logger.Info("Starting processIdleWatches", zap.Int64("token", token))

	playIds := make(map[string]playInfo)
	if token == 0 {
		query := `SELECT DISTINCT token(profile_id) FROM recent_watch LIMIT 1;`

		if err := t.scylla.Session().Query(query).Scan(&token); err != nil {
			if err == gocql.ErrNotFound {
				t.logger.Info("No initial token found, retrying...")
				return nil, token, nil
			}
			t.logger.Error("Error fetching profile ID", zap.Error(err))
			return nil, token, err
		}
	} else {
		query := `SELECT DISTINCT token(profile_id) FROM recent_watch WHERE token(profile_id) > ? LIMIT 1;`
		if err := t.scylla.Session().Query(query, token).Consistency(gocql.One).Scan(&token); err != nil {
			if err == gocql.ErrNotFound {
				t.logger.Info("No next token found, retrying...")
				return nil, token, nil
			}
			t.logger.Error("Error fetching next token", zap.Error(err))
			return nil, token, err
		}
	}

	t.logger.Info("Querying play IDs for token", zap.Int64("token", token))

	query := `SELECT play_id, watched_at, duration, profile_id FROM recent_watch WHERE token(profile_id) = ?;`
	iter := t.scylla.Session().Query(query, token).Consistency(gocql.One).Iter()

	var playId string
	var profile_id string
	var duration int
	var watchedAt gocql.UUID

	for iter.Scan(&playId, &watchedAt, &duration, &profile_id) {
		isOlderThanAgeLimit := watchedAt.Time().Before(daysAgo)
		if !isOlderThanAgeLimit {
			continue
		}

		if watchedAt.Time().Before(daysAgo) {
			existingPlayInfo, exists := playIds[playId]
			if !exists || watchedAt.Time().After(existingPlayInfo.watchedAt.Time()) {
				playIds[playId] = playInfo{
					profile_id: profile_id,
					play_id:    playId,
					watchedAt:  watchedAt,
					duration:   duration,
				}
			}
		}
	}

	if err := iter.Close(); err != nil {
		t.logger.Error("Error fetching play IDs", zap.Error(err))
		return nil, token, err
	}

	t.logger.Info("Successfully processed idle watches", zap.Int("playIdsCount", len(playIds)))
	return playIds, token, nil
}
