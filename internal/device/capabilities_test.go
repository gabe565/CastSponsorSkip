package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vishen/go-chromecast/dns"
)

func TestHasVideoOut(t *testing.T) {
	type args struct {
		entry dns.CastEntry
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr require.ErrorAssertionFunc
	}{
		{"no caps", args{entry: dns.CastEntry{}}, false, require.Error},
		{"invalid caps", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "abc"}}}, false, require.Error},
		{"Google Home", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "199172"}}}, false, require.NoError},
		{"Google Home Mini", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "199428"}}}, false, require.NoError},
		{"Google Nest Mini", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "199172"}}}, false, require.NoError},
		{"Google Nest Hub", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "231941"}}}, true, require.NoError},
		{"Nvidia Shield", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "463365"}}}, true, require.NoError},
		{"Chromecast Ultra", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "201221"}}}, true, require.NoError},
		{"MagniFi Max", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "2052"}}}, false, require.NoError},
		{"BRAVIA 4K 2015", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "264709"}}}, true, require.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HasVideoOut(tt.args.entry)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
