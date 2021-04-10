package torrentfile

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var write = flag.Bool("write", true, "overwrites json files")

func TestOpenFile(t *testing.T) {
	torrentPath := "./testdata/debian-10.9.0-amd64-netinst.iso.torrent"
	jsonPath := "./testdata/debian-10.9.0-amd64-netinst.iso.json"

	torrent, err := OpenFile(torrentPath)
	require.Nil(t, err)

	if *write {
		serialized, err := json.MarshalIndent(torrent, "", "  ")
		require.Nil(t, err)
		ioutil.WriteFile(jsonPath, serialized, 0644)
	}

	expected := TorrentFile{}
	golden, err := ioutil.ReadFile(jsonPath)
	require.Nil(t, err)
	err = json.Unmarshal(golden, &expected)
	require.Nil(t, err)

	fmt.Println(expected, torrent)
	assert.Equal(t, expected, torrent)
}
