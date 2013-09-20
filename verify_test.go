package taipei

import (
	"testing"
)

func TestSingleMode(t *testing.T) {
	m, err := GetMetaInfo("testData/test1.torrent")
	if err != nil {
		t.Errorf(err.Error())
	}
	if v, _ := VerifySingle(m, "testData"); v == false {
		t.Errorf("Verify Content failed.")
	}
}

func TestFullMode(t *testing.T) {
	m, err := GetMetaInfo("testData/test2.torrent")
	if err != nil {
		t.Errorf(err.Error())
	}
	if v, _ := VerifyFull(m, "testData"); v == false {
		t.Errorf("Verify Content failed.")
	}
}

func TestPartialMode1(t *testing.T) {
	m, err := GetMetaInfo("testData/test2.torrent")
	if err != nil {
		t.Errorf(err.Error())
	}
	if v, _ := VerifyPartial(m, "testData"); v == false {
		t.Errorf("Verify Content failed.")
	}
}

func TestPartialMode2(t *testing.T) {
	m, err := GetMetaInfo("testData/test3.torrent")
	if err != nil {
		t.Errorf(err.Error())
	}
	if v, _ := VerifyPartial(m, "testData"); v == false {
		t.Errorf("Verify Content failed.")
	}
}
