package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	tests := map[string]struct {
		sw        *StorageWorker
		index     int
		newIndex  int
		fileIndex int
		fails     bool
	}{
		"correct input": {
			sw:        &StorageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, Queue: nil, Exit: nil},
			index:     1300,
			newIndex:  800,
			fileIndex: 1,
			fails:     false,
		},
		"correct input: on edge": {
			sw:        &StorageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, Queue: nil, Exit: nil},
			index:     1500,
			newIndex:  0,
			fileIndex: 2,
			fails:     false,
		},
		"index out of range": {
			sw:        &StorageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, Queue: nil, Exit: nil},
			index:     5000,
			newIndex:  0,
			fileIndex: 0,
			fails:     true,
		},
	}

	for name, test := range tests {
		index, fileIndex, err := test.sw.GetFile(test.index)
		if test.fails {
			assert.NotNil(t, err, name)
		} else {
			assert.Nil(t, err, name)
			assert.Equal(t, test.newIndex, index, name)
			assert.Equal(t, test.fileIndex, fileIndex, name)
		}
	}
}
