package taipei

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	echo = false
)

func SetEcho(b bool) {
	echo = b
}

func ProgressBar(i, j int) string {
	if i > j {
		i = j
	}
	bs := i * 20 / j
	ps := float32(i) / float32(j) * 100
	ls := float32(i)*20/float32(j) - float32(bs)
	var bf bytes.Buffer
	var k int
	bf.WriteByte('[')
	for k = 0; k < bs; k++ {
		bf.WriteByte('=')
	}
	if ls > 0.5 {
		bf.WriteByte('>')
		k++
	}
	for k < 20 {
		bf.WriteByte(' ')
		k++
	}
	bf.WriteByte(']')
	bf.WriteString(fmt.Sprintf(" %.2f%%", ps))
	return bf.String()
}

func VerifyContent(m *MetaInfo, root string) (bool, error) {
	n := len(m.Info.Files)
	if n == 0 {
		return VerifySingle(m, root)
	}
	for i, _ := range m.Info.Files {
		src := &m.Info.Files[i]
		fullPath := filepath.Join(root, filepath.Clean(filepath.Join(src.Path...)))
		_, err := os.Stat(fullPath)
		if err != nil {
			return VerifyPartial(m, root)
		}
	}
	return VerifyFull(m, root)
}

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
	g, b, err := CheckPieces(fs, size, m)
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
			fs.offsets = append(fs.offsets, size)
			size += src.Length
		}
	}
	g, b, err := CheckPieces(fs, size, m)
	if err != nil {
		return false, err
	}
	log.Println("Good pieces:", g, "Bad pieces:", b)
	return b == 0, nil
}

func VerifyPartial(m *MetaInfo, root string) (bool, error) {
	fs := new(fileStore)
	numFiles := len(m.Info.Files)
	var size int64 = 0
	if numFiles == 0 {
		return false, errors.New("Torrent has single file structure.")
	} else {
		fs.files = make([]fileEntry, 0, numFiles)
		fs.offsets = make([]int64, 0, numFiles)
		var fd *os.File
		for i, _ := range m.Info.Files {
			src := &m.Info.Files[i]
			fullPath := filepath.Join(root, filepath.Clean(filepath.Join(src.Path...)))
			stat, err := os.Stat(fullPath)
			if err != nil {
				log.Println("Skip file:", fullPath)
				fd = nil
			} else {
				if stat.Size() != src.Length {
					return false, errors.New(fullPath + ": size not match.")
				}
				fd, err = os.Open(fullPath)
				if err != nil {
					return false, err
				}
				defer fd.Close()
			}
			fs.files = append(fs.files, fileEntry{src.Length, fd})
			fs.offsets = append(fs.offsets, size)
			size += src.Length
		}
	}
	pieceNum := int((size + m.Info.PieceLength - 1) / m.Info.PieceLength)
	p := 0
	for i := 0; i < pieceNum; i++ {
		g, err := CheckPiece(fs, size, m, i)
		if g == false {
			return false, err
		}
		if err != nil {
			if err == missingPieceErr {
				p++
			} else {
				log.Println(err)
			}
		}
	}
	if p > 0 {
		log.Println(missingPieceErr, ":", p)
	}
	return true, nil
}
