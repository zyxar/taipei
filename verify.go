package taipei

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func VerifySingle(m *MetaInfo, root string) (bool, error) {
	fs := new(fileStore)
	numFiles := len(m.Info.Files)
	var size int64 = 0
	if numFiles == 0 {
		fd, err := os.Open(filepath.Join(root, m.Info.Name)) // TODO: make sure filename is not modified.
		if err != nil {
			return false, err
		}
		defer fd.Close()
		stat, err := fd.Stat()
		if stat.Size() != m.Info.Length {
			return false, errors.New(m.Info.Name + ": size not match.")
		}
		if err != nil {
			return false, err
		}
		fs.files = []fileEntry{{stat.Size(), fd}}
		fs.offsets = []int64{0}
		size = stat.Size()
	} else {
		return false, errors.New("Torrent has multiple file structure.")
	}
	g, b, _, err := CheckPieces(fs, size, m)
	if err != nil {
		return false, err
	}
	log.Println("Good pieces:", g, "Bad pieces:", b)
	return b == 0, nil
}

func VerifyFull(m *MetaInfo, root string) (bool, error) {
	fs := new(fileStore)
	numFiles := len(m.Info.Files)
	var size int64 = 0
	if numFiles == 0 {
		return false, errors.New("Torrent has single file structure.")
	} else {
		fs.files = make([]fileEntry, 0, numFiles)
		fs.offsets = make([]int64, 0, numFiles)
		for i, _ := range m.Info.Files {
			src := &m.Info.Files[i]
			fullPath := filepath.Join(root, filepath.Clean(filepath.Join(src.Path...)))
			stat, err := os.Stat(fullPath)
			if err != nil {
				return false, err
			}
			if stat.Size() != src.Length {
				return false, errors.New(fullPath + ": size not match.")
			}
			fd, err := os.Open(fullPath)
			if err != nil {
				return false, err
			}
			defer fd.Close()
			fs.files = append(fs.files, fileEntry{src.Length, fd})
			fs.offsets = append(fs.offsets, src.Length*int64(i))
			size += src.Length
		}
	}
	g, b, _, err := CheckPieces(fs, size, m)
	if err != nil {
		return false, err
	}
	log.Println("Good pieces:", g, "Bad pieces:", b)
	return b == 0, nil
}
