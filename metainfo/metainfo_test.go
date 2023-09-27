package metainfo

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromFile(t *testing.T) {
	torrentPath := "./testdata/debian-10.9.0-amd64-netinst.iso.torrent"
	jsonPath := "./testdata/debian-10.9.0-amd64-netinst.iso.json"

	info, err := FromFile(torrentPath)
	require.Nil(t, err)

	if false {
		// overwrite expected output json
		serialized, err := json.MarshalIndent(info, "", "  ")
		require.Nil(t, err)
		os.WriteFile(jsonPath, serialized, 0644)
	}

	expected := &MetaInfo{}
	format, err := os.ReadFile(jsonPath)
	require.Nil(t, err)
	err = json.Unmarshal(format, &expected)
	require.Nil(t, err)

	assert.Equal(t, expected, info)
}

func TestToMetaInfo(t *testing.T) {
	tests := map[string]struct {
		input  *bencodeTorrent
		output *MetaInfo
		fails  bool
	}{
		"correct input": {
			input: &bencodeTorrent{
				Announce: "http://test.tracker.org:6969/announce",
				Info: bencodeInfo{
					Pieces:      "T0e1S2t3P4i5E6c7E8s9T0e1S2t3P4i5E6c7E8s9",
					PieceLength: 262144,
					Length:      351272960,
					Name:        "test.iso",
				},
			},
			output: &MetaInfo{
				Announce: []string{"http://test.tracker.org:6969/announce"},
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
		"correct input: announce-list": {
			input: &bencodeTorrent{
				AnnounceList: []string{"http://first.tracker.org:6969/announce", "http://second.tracker.org:6969/announce"},
				Info: bencodeInfo{
					Pieces:      "T0e1S2t3P4i5E6c7E8s9T0e1S2t3P4i5E6c7E8s9",
					PieceLength: 262144,
					Length:      351272960,
					Name:        "test.iso",
				},
			},
			output: &MetaInfo{
				Announce: []string{"http://first.tracker.org:6969/announce", "http://second.tracker.org:6969/announce"},
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
			input: &bencodeTorrent{
				Announce: "http://test.tracker.org:6969/announce",
				Info: bencodeInfo{
					Pieces:      "T1e2S3t4P5i6E7c8E9s10", // <- fails here: only 20 bytes
					PieceLength: 262144,
					Length:      351272960,
					Name:        "test.iso",
				},
			},
			output: nil,
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
