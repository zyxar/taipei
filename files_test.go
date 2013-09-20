package taipei

import (
	"crypto/sha1"
	"fmt"
	"os"
	"testing"
)

type testFile struct {
	path    string
	fileLen int64
	hash    string
	// SHA1 of the first 25 bytes only.
	hashPieceA string
	// SHA1 of bytes 25-49
	hashPieceB string
}

var tests []testFile = []testFile{{
	"testData/test1.zip",
	1024,
	// "A0AD08765665C1339E2F829F4EBFF598B355A62B",
	"76BCDE224CB19164B40AE06B2606A9F583C1C9BE",
	// dd if=testdata/file bs=25 count=1 | shasum
	"194F7B8DD0A5B339440D2397BA5251EBD3A52D5D",
	// dd if=testdata/file bs=25 count=1 skip=1 | shasum
	"51535521A5F1CA1DB3D215A1A5EC57F54B4C844A",
}}

func mkFileStore(tf testFile) (fs *fileStore, err error) {
	fd, err := os.Open(tf.path)
	if err != nil {
		return fs, err
	}
	f := fileEntry{tf.fileLen, fd}
	return &fileStore{[]int64{0}, []fileEntry{f}}, nil
}

func TestFileStoreRead(t *testing.T) {
	for _, testFile := range tests {
		fs, err := mkFileStore(testFile)
		if err != nil {
			t.Fatal(err)
		}
		ret := make([]byte, testFile.fileLen)
		_, err = fs.ReadAt(ret, 0)
		if err != nil {
			t.Fatal(err)
		}
		h := sha1.New()
		h.Write(ret)
		sum := fmt.Sprintf("%X", h.Sum(nil))
		if sum != testFile.hash {
			t.Errorf("Wanted %v, got %v\n", testFile.hash, sum)
		}
	}
}
