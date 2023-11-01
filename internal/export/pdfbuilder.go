package export

import "github.com/chooban/progdl-go/internal/env"

// BuildPdf will export a PDF of the provided series and optional
// episodes.
// The parameters of seriesTitle and episodeTitle should be used to
// query the database via appEnv.Db to retrieve all applicable episodes,
// ordered by issue number.
func BuildPdf(appEnv env.AppEnv, seriesTitle string, episodeTitle string) {

}
