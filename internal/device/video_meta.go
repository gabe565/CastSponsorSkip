package device

type VideoMeta struct {
	CurrVideoId string
	CurrArtist  string
	CurrTitle   string

	PrevVideoId string
	PrevArtist  string
	PrevTitle   string
}

func (v *VideoMeta) Clear() {
	v.CurrVideoId = ""
	v.CurrArtist = ""
	v.CurrTitle = ""
	v.PrevVideoId = ""
	v.PrevArtist = ""
	v.PrevTitle = ""
}

func (v VideoMeta) Empty() bool {
	return v.CurrArtist == "" || v.CurrTitle == ""
}

func (v VideoMeta) SameVideo() bool {
	return v.CurrArtist == v.PrevArtist && v.CurrTitle == v.PrevTitle
}
