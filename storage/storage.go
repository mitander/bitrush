package storage

import (
	"errors"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type File struct {
	Path   string
	Length int
}

type StorageWork struct {
	Data  []byte
	Index int64
}

type StorageWorker struct {
	files       []*os.File
	fileLengths []int
	Queue       chan (StorageWork)
	Exit        chan (int)
}

func NewStorageWorker(dir string, files []File) (*StorageWorker, error) {
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
			return nil, err
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
			log.WithFields(log.Fields{"reason": err.Error(), "path": f.Path}).Error("failed to open file")
			return nil, err
		}
		osFiles = append(osFiles, file)
		fileLengths = append(fileLengths, f.Length)
	}

	return &StorageWorker{
		files:       osFiles,
		fileLengths: fileLengths,
		Queue:       make(chan (StorageWork)),
		Exit:        make(chan (int)),
	}, nil
}

func (s *StorageWorker) StartWorker() {
	for {
		select {
		case w := <-s.Queue:
			index, fileIndex, err := s.GetFile(int(w.Index))
			if err != nil {
				log.WithFields(log.Fields{
					"reason": err.Error(),
				}).Error("failed to get file, putting work back in queue")
				s.Queue <- w
			}
			file := s.files[fileIndex]
			_, err = file.Seek((int64(index)), 0)
			if err != nil {
				log.WithFields(log.Fields{
					"reason": err.Error(),
				}).Error("failed to seek file, putting work back in queue")
				s.Queue <- w
			}
			l, err := file.Write(w.Data)
			if err != nil {
				log.WithFields(log.Fields{
					"reason": err.Error(),
				}).Error("failed writing to file")
				s.Queue <- w
			}

			log.Debugf("wrote %d bytes to index %d", l, w.Index)
			continue
		case <-s.Exit:
			log.Debug("received exit signal, exiting storage worker")
			close(s.Queue)
			close(s.Exit)
			for _, f := range s.files {
				f.Close()
			}
			return
		}
	}
}

func (s *StorageWorker) GetFile(index int) (int, int, error) {
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
