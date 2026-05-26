package models

type EpisodeCreator struct {
	ID        int64  `db:"id"`
	EpisodeID int64  `db:"episode_id"`
	CreatorID int64  `db:"creator_id"`
	Role      string `db:"role"`
}
