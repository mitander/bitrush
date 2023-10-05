package storage

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	tests := map[string]struct {
		worker    *StorageWorker
		index     int
		newIndex  int
		fileIndex int
		fails     bool
	}{
		"correct input": {
			worker:    &StorageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, Queue: nil, ctx: nil},
			index:     1300,
			newIndex:  800,
			fileIndex: 1,
			fails:     false,
		},
		"correct input: on edge": {
			worker:    &StorageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, Queue: nil, ctx: nil},
			index:     1500,
			newIndex:  0,
			fileIndex: 2,
			fails:     false,
		},
		"index out of range": {
			worker:    &StorageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, Queue: nil, ctx: nil},
			index:     5000,
			newIndex:  0,
			fileIndex: 0,
			fails:     true,
		},
	}

	for name, test := range tests {
		index, fileIndex, err := test.worker.GetFile(test.index)
		if test.fails {
			assert.NotNil(t, err, name)
		} else {
			assert.Nil(t, err, name)
			assert.Equal(t, test.newIndex, index, name)
			assert.Equal(t, test.fileIndex, fileIndex, name)
		}
	}
}

func TestSplitFileBounds(t *testing.T) {
	// randomized bytes to also verify split content
	data := make([]byte, 200)
	rand.Read(data)

	tests := map[string]struct {
		worker *StorageWorker
		work   StorageWork
		split  *StorageWork
	}{
		"test 1": {
			worker: &StorageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, Queue: nil, ctx: nil},
			work:   StorageWork{Data: data, Index: 400},
			split:  &StorageWork{Data: data[100:], Index: 500},
		},
		"test 2": {
			worker: &StorageWorker{files: nil, fileLengths: []int{200, 350, 400}, Queue: nil, ctx: nil},
			work:   StorageWork{Data: data, Index: 546},
			split:  &StorageWork{Data: data[4:], Index: 550},
		},
		"test 3": {
			worker: &StorageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, Queue: nil, ctx: nil},
			work:   StorageWork{Data: data, Index: 1300},
			split:  nil,
		},
	}

	for name, test := range tests {
		index, file, err := test.worker.GetFile(test.work.Index)
		assert.Nil(t, err, name)
		split := test.worker.SplitFileBounds(test.work, index, file)
		if test.split == nil {
			assert.Nil(t, split, name)
		} else {
			assert.Equal(t, test.split, split, name)
		}
	}
}
