package pdf

type Bookmark struct {
	Title    string
	PageFrom int
	PageThru int
}

type EpisodeDetails struct {
	Bookmark Bookmark
	Credits  string
}
