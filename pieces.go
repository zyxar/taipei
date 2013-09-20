// Compute missing pieces for a torrent.
package taipei

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"runtime"
)

func CheckPieces(fs FileStore, totalLength int64, m *MetaInfo) (good, bad int, goodBits *Bitset, err error) {
	pieceLength := m.Info.PieceLength
	numPieces := int((totalLength + pieceLength - 1) / pieceLength)
	goodBits = NewBitset(int(numPieces))
	ref := m.Info.Pieces
	if len(ref) != numPieces*sha1.Size {
		err = errors.New("Incorrect Info.Pieces length")
		return
	}
	currentSums, err := ComputeSums(fs, totalLength, m.Info.PieceLength)
	if err != nil {
		return
	}
	for i := 0; i < numPieces; i++ {
		base := i * sha1.Size
		end := base + sha1.Size
		if bytes.Compare([]byte(ref[base:end]), currentSums[base:end]) == 0 {
			good++
			goodBits.Set(int(i))
		} else {
			bad++
		}
	}
	return
}

type chunk struct {
	i    int64
	data []byte
}

// ComputeSums reads the file content and computes the SHA1 hash for each
// piece. Spawns parallel goroutines to compute the hashes, since each
// computation takes ~30ms.
func ComputeSums(fs FileStore, totalLength int64, pieceLength int64) (sums []byte, err error) {
	// Calculate the SHA1 hash for each piece in parallel goroutines.
	hashes := make(chan chunk)
	results := make(chan chunk, 3)
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go hashPiece(hashes, results)
	}

	// Read file content and send to "pieces", keeping order.
	numPieces := (totalLength + pieceLength - 1) / pieceLength
	go func() {
		for i := int64(0); i < numPieces; i++ {
			piece := make([]byte, pieceLength, pieceLength)
			if i == numPieces-1 {
				piece = piece[0 : totalLength-i*pieceLength]
			}
			// Ignore errors.
			fs.ReadAt(piece, i*pieceLength)
			hashes <- chunk{i: i, data: piece}
		}
		close(hashes)
	}()

	// Merge back the results.
	sums = make([]byte, sha1.Size*numPieces)
	for i := int64(0); i < numPieces; i++ {
		h := <-results
		copy(sums[h.i*sha1.Size:], h.data)
	}
	return
}

func hashPiece(h chan chunk, result chan chunk) {
	hasher := sha1.New()
	for piece := range h {
		hasher.Reset()
		_, err := hasher.Write(piece.data)
		if err != nil {
			result <- chunk{piece.i, nil}
		} else {
			result <- chunk{piece.i, hasher.Sum(nil)}
		}
	}
}

func CheckPiece(fs FileStore, totalLength int64, m *MetaInfo, pieceIndex int) (good bool, err error) {
	ref := m.Info.Pieces
	currentSum, err := ComputePieceSum(fs, totalLength, m.Info.PieceLength, pieceIndex)
	if err != nil {
		if err == os.ErrInvalid {
			good = true
			err = errors.New("Missing file for critical piece.")
		}
		return
	}
	base := pieceIndex * sha1.Size
	end := base + sha1.Size
	refSha1 := []byte(ref[base:end])
	good = (bytes.Compare(refSha1, currentSum) == 0)
	if !good {
		err = fmt.Errorf("reference sha1: %v != piece sha1: %v", refSha1, currentSum)
	}
	return
}

func ComputePieceSum(fs FileStore, totalLength int64, pieceLength int64, pieceIndex int) (sum []byte, err error) {
	numPieces := (totalLength + pieceLength - 1) / pieceLength
	hasher := sha1.New()
	piece := make([]byte, pieceLength)
	if int64(pieceIndex) == numPieces-1 {
		piece = piece[0 : totalLength-int64(pieceIndex)*pieceLength]
	}
	_, err = fs.ReadAt(piece, int64(pieceIndex)*pieceLength)
	if err != nil {
		return
	}
	_, err = hasher.Write(piece)
	if err != nil {
		return
	}
	sum = hasher.Sum(nil)
	return
}
