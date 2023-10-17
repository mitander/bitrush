package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

type File struct {
	Path   string
	Length int
}

type storageWork struct {
	Data  []byte
	Index int
}

type storageWorker struct {
	files       []*os.File
	fileLengths []int
	queue       chan (storageWork)
	ctx         context.Context
}

func NewStorageWorker(ctx context.Context, dir string, files []File) (*storageWorker, error) {
	err := os.Mkdir(dir, 0755)
	if err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}

	if len(files) > 1 {
		// root folder
		dir = filepath.Join(dir, files[0].Path)
		err := os.Mkdir(dir, 0755)
		if err != nil {
			if !os.IsExist(err) {
				return nil, err
			}
		}
	}

	var fileLengths []int
	var osFiles []*os.File
	for _, f := range files {
		if f.Length == 0 {
			continue
		}
		path := filepath.Join(dir, f.Path)
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("failed to open filex")
			return nil, err
		}
		osFiles = append(osFiles, file)
		fileLengths = append(fileLengths, f.Length)
	}

	return &storageWorker{
		files:       osFiles,
		fileLengths: fileLengths,
		queue:       make(chan (storageWork)),
		ctx:         ctx,
	}, nil
}

func (s *storageWorker) StartWorker() {
	for {
		select {
		case w := <-s.queue:
			index, fileIndex, err := s.getFile(w.Index)
			if err != nil {
				log.Errorf("putting piece %d back in queue: could not get file", w.Index)
				s.queue <- w
				continue
			}

			split := s.splitFileBounds(w, index, fileIndex)
			if split != nil {
				// piece data overlap file bounds,
				// split rest data to new storage work
				s.queue <- *split
			}

			file := s.files[fileIndex]
			l, err := s.write(file, w)
			if err != nil {
				log.Errorf("putting piece %d back in queue: could not store work", w.Index)
				s.queue <- w
				continue
			}

			log.WithFields(log.Fields{
				"file":   fileIndex,
				"index":  index,
				"length": l,
			}).Debug("wrote to file")

		case <-s.ctx.Done():
			close(s.queue)
			for _, f := range s.files {
				f.Close()
			}
			log.Debug("received exit signal, exiting storage worker")
			return
		}
	}
}

func (s *storageWorker) getFile(index int) (int, int, error) {
	if len(s.files) == 1 {
		return index, 0, nil
	}
	var offset int
	for i, l := range s.fileLengths {
		offset += l
		if index >= offset {
			continue
		}
		idx := index - (offset - l)
		return idx, i, nil
	}
	return 0, 0, errors.New("index not in range")
}

func (s *storageWorker) write(w io.WriteSeeker, sw storageWork) (int, error) {
	_, err := w.Seek((int64(sw.Index)), 0)
	if err != nil {
		log.WithFields(log.Fields{
			"work":   sw,
			"reason": err.Error(),
		}).Error("failed to seek file")
		return 0, err
	}

	l, err := w.Write(sw.Data)
	if err != nil {
		log.WithFields(log.Fields{
			"work":   sw,
			"reason": err.Error(),
		}).Error("failed writing to file")
		return 0, err
	}
	return l, nil
}

func (s *storageWorker) splitFileBounds(w storageWork, index int, fileIndex int) *storageWork {
	end := index + len(w.Data)
	fileLen := s.fileLengths[fileIndex]
	if end > fileLen {
		split := fileLen - index
		restData := w.Data[split:]
		w.Data = w.Data[:split]
		log.WithFields(log.Fields{
			"end":      end,
			"fileLen":  fileLen,
			"split":    split,
			"index":    w.Index,
			"newIndex": w.Index + split,
		}).Debug("split storage work")
		return &storageWork{Index: w.Index + split, Data: restData}
	}
	return nil
}

func (s *storageWorker) Complete() error {
	for len(s.queue) != 0 {
		time.Sleep(1 * time.Second)
		log.Debugf("storage work not completed: %d work items left", len(s.queue))
	}

	ok := true
	for i := range s.files {
		stat, err := s.files[i].Stat()
		if err != nil {
			return err
		}
		if stat.Size() != int64(s.fileLengths[i]) {
			ok = false
		}
	}
	if !ok {
		return errors.New("file lengths not matching")
	}

	log.Debugf("storage work completed")
	return nil
}

func (s *storageWorker) AddWork(data []byte, index int) {
	s.queue <- storageWork{Data: data, Index: index}
}
