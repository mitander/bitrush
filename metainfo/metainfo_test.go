package metainfo

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mitander/bitrush/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetaInfo(t *testing.T) {
	torrentPath := "./testdata/debian-10.9.0-amd64-netinst.iso.torrent"
	jsonPath := "./testdata/debian-10.9.0-amd64-netinst.iso.json"

	info, err := NewMetaInfo(torrentPath)
	require.Nil(t, err)

	if true {
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
					Files:       []bencodeFile{},
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
				Files:       []storage.File{{Path: "test.iso", Length: 351272960}},
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
					Files:       []bencodeFile{},
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
				Files:       []storage.File{{Path: "test.iso", Length: 351272960}},
			},
			fails: false,
		},
		"correct input: multi-file": {
			input: &bencodeTorrent{
				Announce: "http://test.tracker.org:6969/announce",
				Info: bencodeInfo{
					Pieces:      "T0e1S2t3P4i5E6c7E8s9T0e1S2t3P4i5E6c7E8s9",
					PieceLength: 262144,
					Length:      351272960,
					Name:        "MultiFileDownload",
					Files:       []bencodeFile{{Path: []string{"notes.txt"}, Length: 2000}, {Path: []string{"passwords.txt"}, Length: 1300}},
				},
			},
			output: &MetaInfo{
				Announce: []string{"http://test.tracker.org:6969/announce"},
				InfoHash: [20]byte{181, 195, 58, 171, 46, 218, 137, 221, 88, 101, 209, 27, 104, 249, 234, 211, 106, 49, 4, 92},
				PieceHashes: [][20]byte{
					{84, 48, 101, 49, 83, 50, 116, 51, 80, 52, 105, 53, 69, 54, 99, 55, 69, 56, 115, 57},
					{84, 48, 101, 49, 83, 50, 116, 51, 80, 52, 105, 53, 69, 54, 99, 55, 69, 56, 115, 57},
				},
				PieceLength: 262144,
				Length:      3300,
				Name:        "MultiFileDownload",
				Files:       []storage.File{{Path: "MultiFileDownload", Length: 0}, {Path: "notes.txt", Length: 2000}, {Path: "passwords.txt", Length: 1300}},
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
					Files:       []bencodeFile{{Path: []string{"test.iso"}, Length: 351272960}},
				},
			},
			output: nil,
			fails:  true,
		},
	}

	for name, test := range tests {
		tf, err := test.input.toMetaInfo()
		if test.fails {
			assert.NotNil(t, err, name)
		} else {
			assert.Nil(t, err, name)
		}
		assert.Equal(t, test.output, tf, name)
	}
}
