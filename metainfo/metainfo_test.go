package metainfo

import (
	"encoding/json"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var write = flag.Bool("write", true, "overwrites json files")

func TestOpenFile(t *testing.T) {
	torrentPath := "./testdata/debian-10.9.0-amd64-netinst.iso.torrent"
	jsonPath := "./testdata/debian-10.9.0-amd64-netinst.iso.json"

	info, err := OpenFile(torrentPath)
	require.Nil(t, err)

	if *write {
		serialized, err := json.MarshalIndent(info, "", "  ")
		require.Nil(t, err)
		os.WriteFile(jsonPath, serialized, 0644)
	}

	expected := MetaInfo{}
	format, err := os.ReadFile(jsonPath)
	require.Nil(t, err)
	err = json.Unmarshal(format, &expected)
	require.Nil(t, err)

	assert.Equal(t, expected, info)
}

func TestToMetaInfo(t *testing.T) {
	tests := map[string]struct {
		input  *torrentBencode
		output MetaInfo
		fails  bool
	}{
		"correct input": {
			input: &torrentBencode{
				Announce: "http://test.tracker.org:6969/announce",
				Info: infoBencode{
					Pieces:      "T0e1S2t3P4i5E6c7E8s9T0e1S2t3P4i5E6c7E8s9",
					PieceLength: 262144,
					Length:      351272960,
					Name:        "test.iso",
				},
			},
			output: MetaInfo{
				Announce: "http://test.tracker.org:6969/announce",
				InfoHash: [20]byte{148, 102, 213, 85, 174, 246, 146, 126, 127, 246, 85, 15, 22, 6, 186, 128, 220, 105, 12, 15},
				PieceHashes: [][20]byte{
					{84, 48, 101, 49, 83, 50, 116, 51, 80, 52, 105, 53, 69, 54, 99, 55, 69, 56, 115, 57},
					{84, 48, 101, 49, 83, 50, 116, 51, 80, 52, 105, 53, 69, 54, 99, 55, 69, 56, 115, 57},
				},
				PieceLength: 262144,
				Length:      351272960,
				Name:        "test.iso",
			},
			fails: false,
		},
		"invalid pieces length": {
			input: &torrentBencode{
				Announce: "http://test.tracker.org:6969/announce",
				Info: infoBencode{
					Pieces:      "T1e2S3t4P5i6E7c8E9s10", // <- fails here: only 20 bytes
					PieceLength: 262144,
					Length:      351272960,
					Name:        "test.iso",
				},
			},
			output: MetaInfo{},
			fails:  true,
		},
	}

	for _, test := range tests {
		tf, err := test.input.toMetaInfo()
		if test.fails {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, tf, test.output)
	}
}
