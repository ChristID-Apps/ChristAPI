package bible

type BibleVersion struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type BibleBook struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Testament string `json:"testament"`
	BookOrder int    `json:"book_order"`
}

type BibleVerse struct {
	ID      int64  `json:"id"`
	Chapter int    `json:"chapter"`
	Verse   int    `json:"verse"`
	Text    string `json:"text"`
}

type BibleChapter struct {
	Version BibleVersion `json:"version"`
	Book    BibleBook    `json:"book"`
	Chapter int          `json:"chapter"`
	Verses  []BibleVerse `json:"verses"`
}

type BibleSearchResult struct {
	Version BibleVersion `json:"version"`
	Book    BibleBook    `json:"book"`
	Chapter int          `json:"chapter"`
	Verse   int          `json:"verse"`
	Text    string       `json:"text"`
}

type BibleSearchFilter struct {
	VersionCode string
	BookCode    string
	Testament   string
	Query       string
	Limit       int
	Offset      int
}
