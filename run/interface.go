package run

type container struct {
	Data data `json:"MediaContainer"`
}

type data struct {
	Media   []PlexMedia `json:"Metadata"`
	Library []Library   `json:"Directory"`
}

type PlexMedia struct {
	RatingKey        int64  `json:"ratingKey,string"`
	Key              string `json:"key"`
	Type             string `json:"type"`
	Title            string `json:"title"`
	ParentTitle      string `json:"parentTitle"`
	GrandparentTitle string `json:"grandparentTitle"`
	OriginalTitle    string `json:"originalTitle"`
	EpisodeNumber    int64  `json:"index"`
	SeasonNumber     int64  `json:"parentIndex"`
	ViewOffset       int64  `json:"viewOffset"`
	ViewCount        int64  `json:"viewCount"`
}

type Library struct {
	Key   string `json:"key"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type Queue struct {
	Watched []int64
	FileID  []int64
}
