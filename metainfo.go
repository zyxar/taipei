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

	"github.com/jackpal/bencode-go"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
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
	dec := getDecoder(in.Encoding)
	if dec == nil {
		return in
	}
	if v, err := dec.String(in.Info.Name); err == nil {
		in.Info.Name = v
	}
	for i, _ := range in.Info.Files {
		for j, _ := range in.Info.Files[i].Path {
			if v, err := dec.String(in.Info.Files[i].Path[j]); err == nil {
				in.Info.Files[i].Path[j] = v
			}
		}
	}
	return in
}

func getDecoder(e string) *encoding.Decoder {
	switch strings.ToLower(e) {
	case "gbk":
		return simplifiedchinese.GBK.NewDecoder()
	case "gb2312":
		return simplifiedchinese.HZGB2312.NewDecoder()
	case "gb18030":
		return simplifiedchinese.GB18030.NewDecoder()
	case "big5", "csbig5":
		return traditionalchinese.Big5.NewDecoder()
	case "eucjp", "euc-jp", "extended_unix_code_packed_format_for_japanese", "cseucpkdfmtjapanese":
		return japanese.EUCJP.NewDecoder()
	case "iso2022jp", "iso-2022-jp":
		return japanese.ISO2022JP.NewDecoder()
	case "shiftjis", "shift-jis", "shift_jis", "ms_kanji", "csshiftjis", "sjis", "ibm-943", "windows-31j", "cp932", "windows-932":
		return japanese.ShiftJIS.NewDecoder()
	case "euckr", "euc-kr", "ibm-1363", "ks_c_5601-1987", "ks_c_5601-1989", "ksc_5601", "korean", "iso-ir-149", "cp1363", "5601", "ksc", "windows-949", "ibm-970", "cp970", "970", "cp949":
		return korean.EUCKR.NewDecoder()
	case "utf-16", "utf16":
		return unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
	case "utf-16le", "utf16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
	case "utf-16be", "utf16be":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()
	case "macintosh", "macos-0_2-10.2", "mac", "csmacintosh", "windows-10000", "macroman":
		return charmap.Macintosh.NewDecoder()
	case "iso-8859-1", "latin-1", "latin1", "iso latin 1", "ibm819", "cp819", "iso_8859-1:1987", "iso-ir-100", "l1", "csisolatin1":
		return charmap.ISO8859_1.NewDecoder()
	case "iso-8859-2", "iso88592", "iso_8859-2:1987", "iso-ir-101", "latin2", "latin-2", "l2", "csisolatin2":
		return charmap.ISO8859_2.NewDecoder()
	case "iso-8859-3", "iso88593", "iso_8859-3:1988", "iso-ir-109", "latin3", "latin-3", "l3", "csisolatin3":
		return charmap.ISO8859_3.NewDecoder()
	case "iso-8859-4", "iso88594", "iso_8859-4:1988", "iso-ir-110", "latin4", "latin-4", "l4", "csisolatin4":
		return charmap.ISO8859_4.NewDecoder()
	case "iso-8859-5", "iso88595", "iso_8859-5:1988", "iso-ir-144", "cyrillic", "csisolatincyrillic":
		return charmap.ISO8859_5.NewDecoder()
	case "iso-8859-6", "iso88596", "iso_8859-6:1987", "iso-ir-127", "ecma-114", "asmo-708", "arabic", "csisolatinarabic":
		return charmap.ISO8859_6.NewDecoder()
	case "iso-8859-6i", "iso88596i":
		return charmap.ISO8859_6I.NewDecoder()
	case "iso-8859-6e", "iso88596e":
		return charmap.ISO8859_6E.NewDecoder()
	case "iso-8859-7", "iso88597", "iso_8859-7:2003", "iso-ir-126", "elot_928", "ecma-118", "greek", "greek8", "csisolatingreek":
		return charmap.ISO8859_7.NewDecoder()
	case "iso-8859-8", "iso88598", "iso_8859-8:1999", "iso-ir-138", "hebrew", "csisolatinhebrew":
		return charmap.ISO8859_8.NewDecoder()
	case "iso-8859-8i", "iso88598i":
		return charmap.ISO8859_8I.NewDecoder()
	case "iso-8859-8e", "iso88598e":
		return charmap.ISO8859_8E.NewDecoder()
	case "iso-8859-9", "iso88599", "iso_8859-9:1999", "iso-ir-148", "latin-5", "latin5", "l5", "csisolatin5":
		return nil
	case "iso-8859-10", "iso885910", "iso_8859-10:1992", "l6", "iso-ir-157", "latin6", "latin-6", "csisolatin6":
		return charmap.ISO8859_10.NewDecoder()
	case "iso-8859-11", "iso885911", "iso_8859-11:2001", "latin/thai", "tis-620":
		return nil
	case "iso-8859-13", "iso885913", "latin7", "latin-7", "baltic rim":
		return charmap.ISO8859_13.NewDecoder()
	case "iso-8859-14", "iso885914", "iso-ir-199", "iso_8859-14:1998", "latin8", "iso-celtic", "l8", "latin-8":
		return charmap.ISO8859_14.NewDecoder()
	case "iso-8859-15", "iso885915", "latin-9", "latin9":
		return charmap.ISO8859_15.NewDecoder()
	case "iso-8859-16", "iso885916", "iso-ir-226", "iso_8859-16:2001", "latin10", "l10", "latin-10":
		return charmap.ISO8859_16.NewDecoder()
	case "koi8-r", "koi8r", "cskoi8r":
		return charmap.KOI8R.NewDecoder()
	case "koi8-u", "koi8u":
		return charmap.KOI8U.NewDecoder()
	case "windows-1250", "windows1250":
		return charmap.Windows1250.NewDecoder()
	case "windows-1251", "windows1251", "cp1251":
		return charmap.Windows1251.NewDecoder()
	case "windows-1252", "windows1252", "cp1252":
		return charmap.Windows1252.NewDecoder()
	case "windows-1253", "windows1253":
		return charmap.Windows1253.NewDecoder()
	case "windows-1254", "windows1254":
		return charmap.Windows1254.NewDecoder()
	case "windows-1255", "windows1255":
		return charmap.Windows1255.NewDecoder()
	case "windows-1256", "windows1256":
		return charmap.Windows1256.NewDecoder()
	case "windows-1257", "windows1257":
		return charmap.Windows1257.NewDecoder()
	case "windows-1258", "windows1258":
		return charmap.Windows1258.NewDecoder()
	case "windows-874", "windows874":
		return charmap.Windows874.NewDecoder()
	case "ascii", "us", "us-ascii", "iso646-us", "ibm367", "cp367", "ansi_x3.4-1968", "iso-ir-6", "ansi_x3.4-1986", "iso_646.irv:1991", "csascii":
		return nil
	case "macintoshcyrillic", "macintosh-cyrillic", "macos-7_3-10.2", "x-mac-cyrillic", "windows-10007", "mac-cyrillic", "maccy":
		return charmap.MacintoshCyrillic.NewDecoder()
	default:
	}
	return nil
}
