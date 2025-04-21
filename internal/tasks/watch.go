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

	for {
		// Process idle watches and retrieve play IDs that need updates
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
		wg.Add(1)
		t.workerpool.Submit(func() {
			t.updateWatches(playInfos, profileToken)
			wg.Done()
		})
	}
}

func (t *task) updateWatches(playInfos map[string]playInfo, profileToken int64) {
	batch := t.scylla.Session().NewBatch(gocql.LoggedBatch)
	deleteTS := time.Now().UnixNano() / 1000 // this is for deleteing recent_watches
	batch.SetConsistency(gocql.Any)

	// Build a comma-separated placeholder string for play IDs
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
		return
	}
	var watchedAt gocql.UUID
	var play_id string
	// Query watched table for play IDs and their watched_at timestamps

	// TODO: better query style
	query := fmt.Sprintf(`SELECT play_id, watched_at FROM watched WHERE token(profile_id) = ? AND play_id in (%v)`, placeHolder)
	iter := t.scylla.Session().Query(query, profileToken).Consistency(gocql.One).Iter()
	for iter.Scan(&play_id, &watchedAt) {
		playInfo, ok := playInfos[play_id]

		if !ok {
			t.logger.Error("Play info not found", zap.String("play_id", play_id), zap.String("profile_id", playInfo.profile_id))
			continue
		}
		// watched queries
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
			//delete outdated ordered watched query
			batch.Query(queryDeleteOutdatedOrderedWatch, playInfo.profile_id, watchedAt)
		}
		batch.Query(queryDeleteOutdatedRecentWatch, deleteTS, playInfo.profile_id, playInfo.play_id)

	}
	// Execute the batch
	if err := t.scylla.Session().ExecuteBatch(batch); err != nil {
		fmt.Println(placeHolder)
		fmt.Println("Error executing batch:", err)
		return
	}

	// Close the iterator and handle errors
	if err := iter.Close(); err != nil {
		fmt.Println("Error closing iterator:", err)
		return
	}

}

// procecssIdleWatches processes idle watches by querying the database for
// profile IDs and their associated play IDs that meet specific criteria.
func (t *task) processIdleWatches(token int64, daysAgo time.Time) (map[string]playInfo, int64, error) {
	playIds := make(map[string]playInfo) // Initialize the map to avoid nil dereference
	if token == 0 {
		query := `SELECT DISTINCT token(profile_id) FROM recent_watch LIMIT 1;`

		// Fetch the initial token and profile ID
		if err := t.scylla.Session().Query(query).Scan(&token); err != nil {
			if err == gocql.ErrNotFound {
				t.logger.Info("No Initial token found, retrying...")
				return nil, token, nil
			}
			t.logger.Error("Error fetching profile ID", zap.Error(err))
			return nil, token, err
		}

	} else {
		query := `SELECT DISTINCT token(profile_id) FROM recent_watch WHERE token(profile_id) > ? LIMIT 1;`
		// Fetch the next token value greater than the provided token
		if err := t.scylla.Session().Query(query, token).Consistency(gocql.One).Scan(&token); err != nil {
			if err == gocql.ErrNotFound {
				t.logger.Info("No next token found, retrying...")
				return nil, token, nil
			}
			t.logger.Error("Error fetching next token", zap.Error(err))
			return nil, token, err
		}
	}
	// Query play IDs, watched_at timestamps, and durations for the given token
	query := `SELECT play_id, watched_at, duration, profile_id FROM recent_watch WHERE token(profile_id) = ?;`
	iter := t.scylla.Session().Query(query, token).Consistency(gocql.One).Iter()

	var playId string
	var profile_id string
	var duration int
	var watchedAt gocql.UUID

	for iter.Scan(&playId, &watchedAt, &duration, &profile_id) {
		// Skip play IDs with watched_at timestamps not older than three days
		isOlderThanAgeLimit := watchedAt.Time().Before(daysAgo)
		if !isOlderThanAgeLimit {
			continue
		}

		// Keep only the most recent play IDs with watched_at timestamps at least 3 days old
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
	// Close the iterator and log errors if any
	if err := iter.Close(); err != nil {
		t.logger.Error("Error fetching play IDs", zap.Error(err))
		return nil, token, err
	}

	return playIds, token, nil
}
