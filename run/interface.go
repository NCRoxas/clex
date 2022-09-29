package run

type container struct {
	Data data `json:"MediaContainer"`
}

type data struct {
	Media   []PlexMedia `json:"Metadata"`
	Library []Library   `json:"Directory"`
}

type Library struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type PlexMedia struct {
	RatingKey        int64  `json:"ratingKey,string"`
	Key              string `json:"key"`
	Type             string `json:"type"`
	Title            string `json:"title"`
	ParentTitle      string `json:"parentTitle"`
	GrandparentTitle string `json:"grandparentTitle"`
	EpisodeNumber    int64  `json:"index"`
	SeasonNumber     int64  `json:"parentIndex"`
	ViewOffset       int64  `json:"viewOffset"`
	ViewCount        int64  `json:"viewCount"`
}

type Marked struct {
	Watched      []int64
	EpisodeFiles []int64
}
