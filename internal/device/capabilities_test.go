package device

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vishen/go-chromecast/dns"
)

func TestHasVideoOut(t *testing.T) {
	type args struct {
		entry dns.CastEntry
	}
	tests := []struct {
		name         string
		args         args
		want         bool
		errAssertion assert.ErrorAssertionFunc
	}{
		{"no caps", args{entry: dns.CastEntry{}}, false, assert.Error},
		{"invalid caps", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "abc"}}}, false, assert.Error},
		{"Google Home", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "199172"}}}, false, assert.NoError},
		{"Google Home Mini", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "199428"}}}, false, assert.NoError},
		{"Google Nest Mini", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "199172"}}}, false, assert.NoError},
		{"Google Nest Hub", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "231941"}}}, true, assert.NoError},
		{"Nvidia Shield", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "463365"}}}, true, assert.NoError},
		{"Chromecast Ultra", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "201221"}}}, true, assert.NoError},
		{"MagniFi Max", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "2052"}}}, false, assert.NoError},
		{"BRAVIA 4K 2015", args{entry: dns.CastEntry{InfoFields: map[string]string{"ca": "264709"}}}, true, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HasVideoOut(tt.args.entry)
			tt.errAssertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
