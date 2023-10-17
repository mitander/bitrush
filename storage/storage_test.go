package storage

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	tests := map[string]struct {
		worker    *storageWorker
		index     int
		newIndex  int
		fileIndex int
		fails     bool
	}{
		"correct input": {
			worker:    &storageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, queue: nil, ctx: nil},
			index:     1300,
			newIndex:  800,
			fileIndex: 1,
			fails:     false,
		},
		"correct input: on edge": {
			worker:    &storageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, queue: nil, ctx: nil},
			index:     1500,
			newIndex:  0,
			fileIndex: 2,
			fails:     false,
		},
		"index out of range": {
			worker:    &storageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, queue: nil, ctx: nil},
			index:     5000,
			newIndex:  0,
			fileIndex: 0,
			fails:     true,
		},
	}

	for name, test := range tests {
		index, fileIndex, err := test.worker.getFile(test.index)
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
		worker *storageWorker
		work   storageWork
		split  *storageWork
	}{
		"test 1": {
			worker: &storageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, queue: nil, ctx: nil},
			work:   storageWork{Data: data, Index: 400},
			split:  &storageWork{Data: data[100:], Index: 500},
		},
		"test 2": {
			worker: &storageWorker{files: nil, fileLengths: []int{200, 350, 400}, queue: nil, ctx: nil},
			work:   storageWork{Data: data, Index: 546},
			split:  &storageWork{Data: data[4:], Index: 550},
		},
		"test 3": {
			worker: &storageWorker{files: nil, fileLengths: []int{500, 1000, 2000}, queue: nil, ctx: nil},
			work:   storageWork{Data: data, Index: 1300},
			split:  nil,
		},
	}

	for name, test := range tests {
		index, file, err := test.worker.getFile(test.work.Index)
		assert.Nil(t, err, name)
		split := test.worker.splitFileBounds(test.work, index, file)
		if test.split == nil {
			assert.Nil(t, split, name)
		} else {
			assert.Equal(t, test.split, split, name)
		}
	}
}
