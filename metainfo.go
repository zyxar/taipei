package taipei

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"bencode"
	"github.com/zyxar/mahonia"
)

type FileDict struct {
	Length int64
	Path   []string
	Md5sum string
}

func (f FileDict) String() string {
	return fmt.Sprintf("%d\r\t\t%v\r\t\t\t\t%X", f.Length, filepath.Join(f.Path...), f.Md5sum)
}

type InfoDict struct {
	PieceLength int64 "piece length"
	Pieces      string
	Private     int64
	Name        string
	// Single File Mode
	Length int64
	Md5sum string
	// Multiple File mode
	Files []FileDict
}

func (i InfoDict) String() string {
	if len(i.Files) == 0 {
		return fmt.Sprintf("Name: %v\tPieceLength: %d\nSize: %d", i.Name, i.PieceLength, i.Length)
	}
	r := fmt.Sprintf("Name: %v\tPieceLength: %d\nSize\r\t\tFilename", i.Name, i.PieceLength)
	for j, _ := range i.Files {
		r += fmt.Sprintf("\n%v", i.Files[j])
	}
	return r
}

type MetaInfo struct {
	Info         InfoDict
	InfoHash     string
	Announce     string
	CreationDate string "creation date"
	Comment      string
	CreatedBy    string "created by"
	Encoding     string
}

func (m MetaInfo) String() string {
	return fmt.Sprintf("%v\n%X\t%s", m.Info, m.InfoHash, m.Encoding)
}

func getString(m map[string]interface{}, k string) string {
	if v, ok := m[k]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func GetMetaInfo(torrent string) (metaInfo *MetaInfo, err error) {
	var input io.ReadCloser
	if input, err = os.Open(torrent); err != nil {
		return
	}

	// We need to calcuate the sha1 of the Info map, including every value in the
	// map. The easiest way to do this is to read the data using the Decode
	// API, and then pick through it manually.
	var m interface{}
	m, err = bencode.Decode(input)
	input.Close()
	if err != nil {
		err = errors.New("Couldn't parse torrent file phase 1: " + err.Error())
		return
	}

	topMap, ok := m.(map[string]interface{})
	if !ok {
		err = errors.New("Couldn't parse torrent file phase 2.")
		return
	}

	infoMap, ok := topMap["info"]
	if !ok {
		err = errors.New("Couldn't parse torrent file. info")
		return
	}
	var b bytes.Buffer
	if err = bencode.Marshal(&b, infoMap); err != nil {
		return
	}
	hash := sha1.New()
	hash.Write(b.Bytes())

	var m2 MetaInfo
	err = bencode.Unmarshal(&b, &m2.Info)
	if err != nil {
		return
	}

	m2.InfoHash = string(hash.Sum(nil))
	m2.Announce = getString(topMap, "announce")
	m2.CreationDate = getString(topMap, "creation date")
	m2.Comment = getString(topMap, "comment")
	m2.CreatedBy = getString(topMap, "created by")
	m2.Encoding = getString(topMap, "encoding")

	metaInfo = &m2
	return
}

func DecodeMetaInfo(p []byte) (metaInfo *MetaInfo, err error) {
	input := bytes.NewReader(p)
	// We need to calcuate the sha1 of the Info map, including every value in the
	// map. The easiest way to do this is to read the data using the Decode
	// API, and then pick through it manually.
	var m interface{}
	m, err = bencode.Decode(input)
	if err != nil {
		err = errors.New("Couldn't parse torrent file phase 1: " + err.Error())
		return
	}

	topMap, ok := m.(map[string]interface{})
	if !ok {
		err = errors.New("Couldn't parse torrent file phase 2.")
		return
	}

	infoMap, ok := topMap["info"]
	if !ok {
		err = errors.New("Couldn't parse torrent file. info")
		return
	}
	var b bytes.Buffer
	if err = bencode.Marshal(&b, infoMap); err != nil {
		return
	}
	hash := sha1.New()
	hash.Write(b.Bytes())

	var m2 MetaInfo
	err = bencode.Unmarshal(&b, &m2.Info)
	if err != nil {
		return
	}

	m2.InfoHash = string(hash.Sum(nil))
	m2.Announce = getString(topMap, "announce")
	m2.CreationDate = getString(topMap, "creation date")
	m2.Comment = getString(topMap, "comment")
	m2.CreatedBy = getString(topMap, "created by")
	m2.Encoding = getString(topMap, "encoding")

	metaInfo = &m2
	return
}

func Iconv(in *MetaInfo) *MetaInfo {
	if len(in.Encoding) == 0 || strings.EqualFold(in.Encoding, "UTF-8") {
		return in
	}
	dec := mahonia.NewDecoder(in.Encoding)
	if v, ok := dec.ConvertStringOK(in.Info.Name); ok {
		in.Info.Name = v
	}
	for i, _ := range in.Info.Files {
		for j, _ := range in.Info.Files[i].Path {
			if v, ok := dec.ConvertStringOK(in.Info.Files[i].Path[j]); ok {
				in.Info.Files[i].Path[j] = v
			}
		}
	}
	return in
}
