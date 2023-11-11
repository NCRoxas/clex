package run

type container struct {
	Data data `json:"MediaContainer,omitempty"`
}

type data struct {
	Media   []PlexMedia `json:"Metadata,omitempty"`
	Library []Library   `json:"Directory,omitempty"`
}

type PlexMedia struct {
	RatingKey        int64  `json:"ratingKey,string,omitempty"`
	Key              string `json:"key,omitempty"`
	Type             string `json:"type,omitempty"`
	Title            string `json:"title,omitempty"`
	ParentTitle      string `json:"parentTitle,omitempty"`
	GrandparentTitle string `json:"grandparentTitle,omitempty"`
	OriginalTitle    string `json:"originalTitle,omitempty"`
	EpisodeNumber    int64  `json:"index,omitempty"`
	SeasonNumber     int64  `json:"parentIndex,omitempty"`
	ViewOffset       int64  `json:"viewOffset,omitempty"`
	ViewCount        int64  `json:"viewCount,omitempty"`
}

type Library struct {
	Key   string `json:"key,omitempty"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}

type Queue struct {
	Watched []int64
	FileID  []int64
}
