package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	os.Setenv("TTT_TAGS", "foo,bar,baz")
	cfg := loadConfig()

	testCases := []struct {
		Name     string
		Content  string
		WantName string
		WantUrl  string
		WantNote string
		WantOut  string
	}{
		{
			Name:     "Normal",
			Content:  "[hck](https://github.com/sstadick/hck) And a note",
			WantName: "hck",
			WantUrl:  "https://github.com/sstadick/hck",
			WantNote: "And a note #foo #bar #baz",
			WantOut:  "[hck](https://github.com/sstadick/hck) And a note #foo #bar #baz",
		},
		{
			Name:     "ParseFail",
			Content:  "Some stuff",
			WantName: "Some stuff #foo #bar #baz",
			WantUrl:  "",
			WantNote: "",
			WantOut:  "Some stuff #foo #bar #baz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var task = Task{Content: tc.Content}
			task.Parse(cfg)
			assert.Equal(t, tc.WantName, task.Name)
			assert.Equal(t, tc.WantUrl, task.Url)
			assert.Equal(t, tc.WantNote, task.Note)
			assert.Equal(t, tc.WantOut, task.Out())
		})
	}
}

func TestMain(m *testing.M) {
	// Skip log messages during testing
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}
