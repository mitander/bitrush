package storage

import (
	"os"

	log "github.com/sirupsen/logrus"
)

type StorageWork struct {
	Data  []byte
	Index int64
}

type StorageWorker struct {
	file  *os.File
	Queue chan (StorageWork)
	Exit  chan (int)
}

func NewStorageWorker(path string) (*StorageWorker, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		log.WithFields(log.Fields{"reason": err.Error(), "path": path}).Error("failed to open file")
		return nil, err
	}

	return &StorageWorker{
		file:  file,
		Queue: make(chan (StorageWork)),
		Exit:  make(chan (int)),
	}, nil
}

func (s *StorageWorker) StartWorker() {
	for {
		select {
		case w := <-s.Queue:
			_, err := s.file.Seek(w.Index, 0)
			if err != nil {
				log.WithFields(log.Fields{
					"reason": err.Error(),
				}).Error("failed to seek file, putting work back in queue")
				s.Queue <- w
			}
			l, err := s.file.Write(w.Data)
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
			_ = s.file.Close()
			return
		}
	}
}
