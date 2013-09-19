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
	"testData/a.torrent",
	2374,
	// "A0AD08765665C1339E2F829F4EBFF598B355A62B",
	"A5B4170D1BF93CDDDADD0D774B74F67ED8DCD6E1",
	// dd if=testdata/file bs=25 count=1 | shasum
	"673E4A9C2160110032CF7E4E630F04FDB3AFD6B7",
	// dd if=testdata/file bs=25 count=1 skip=1 | shasum
	"7FF36FA218F99E6D847EBEFD1E58E48E5DBA0C56",
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
