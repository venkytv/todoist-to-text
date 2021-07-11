package main

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestPostToInstapaper(t *testing.T) {
	client := resty.New()
	cfg := loadConfig()

	testCases := []struct {
		Name    string
		Content string
		WantUrl string
	}{
		{
			Name:    "Normal",
			Content: "[hck](https://github.com/sstadick/hck)",
			WantUrl: "https://www.instapaper.com/read/1427468257",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			bookmark, err := postToInstapaper(cfg, client, tc.Content)
			assert.Nil(t, err)
			assert.Equal(t, tc.WantUrl, bookmark)
		})
	}
}
