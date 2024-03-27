package device

type VideoMeta struct {
	CurrVideoID string
	CurrArtist  string
	CurrTitle   string

	PrevVideoID string
	PrevArtist  string
	PrevTitle   string
}

func (v *VideoMeta) Clear() {
	v.CurrVideoID = ""
	v.CurrArtist = ""
	v.CurrTitle = ""
	v.PrevVideoID = ""
	v.PrevArtist = ""
	v.PrevTitle = ""
}

func (v VideoMeta) Empty() bool {
	return v.CurrArtist == "" || v.CurrTitle == ""
}

func (v VideoMeta) SameVideo() bool {
	return v.CurrArtist == v.PrevArtist && v.CurrTitle == v.PrevTitle
}
