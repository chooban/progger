package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAdvertPageText(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		text  string
		advert bool
	}{
		{
			name:   "on sale now lowercase",
			text:   "on sale now at all good newsagents",
			advert: true,
		},
		{
			name:   "on sale now uppercase",
			text:   "ON SALE NOW AT ALL GOOD NEWSAGENTS",
			advert: true,
		},
		{
			name:   "on sale with date pattern",
			text:   "NEXT PROG — on sale 12 March 2025",
			advert: true,
		},
		{
			name:   "on sale with single digit day",
			text:   "on sale 5 January 2024",
			advert: true,
		},
		{
			name:   "isbn pattern",
			text:   "ISBN: 123-4-56789-012-3",
			advert: true,
		},
		{
			name:   "story text no match",
			text:   "Judge Dredd was on patrol in Mega-City One when...",
			advert: false,
		},
		{
			name:   "empty string",
			text:   "",
			advert: false,
		},
		{
			name:   "partial match only - 'sale' without 'on sale now'",
			text:   "Big sale today at the market!",
			advert: false,
		},
		{
			name:   "partial match only - 'sale' without 'on sale now'",
			text:   "Big sale today at the market!",
			advert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.advert, IsAdvertPageText(tt.text))
		})
	}
}
