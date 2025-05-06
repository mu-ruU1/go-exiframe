package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dsoprea/go-exif/v3"
	"github.com/dsoprea/go-exif/v3/common"
)

/*
# Exif

## FocalLengthIn35mmFilm

このタグは 35mm フィルムカメラに換算した焦点距離の値を示す。単位は mm である。記録値が 0 の
場合は焦点距離不明を表す。レンズ焦点距離(FocalLength)タグとは異なるので注意する。

## FocalLength

撮影レンズの実焦点距離を示す。単位は mm である。35mm フィルムカメラの焦点距離には換算しな
い。

## DateTimeOriginal

画像が撮影された日付と時間を記載する。フォーマットは“YYYY:MM:DD HH:MM:SS”。時間は 24 時
間表示し、日付と時間の間に空白文字［20.H］を 1 つ埋める。日時不明の場合は、コロン“:”以外の日
付・時間の文字部を空白文字で埋めるか、または、全てを空白文字で埋めるべきである。文字列の長
さは、NULL を含み 20Byte である。記載が無いときは不明として扱う。

*/

// CIPA DC-008-2024 (Exif3.0)
// https://www.cipa.jp/j/std/std-sec.html
type ExifData struct {
	make      string // カメラメーカー [TAG=0x010f]
	model     string // カメラモデル [TAG=0x0110]
	lensMake  string // レンズメーカー [TAG=0xa433]
	lensModel string // レンズモデル [TAG=0xa434]

	exposureTime            int // 露出時間(シャッタースピード) [TAG=0x829a]
	fNumber                 int // Fナンバー(f値) [TAG=0x829d]
	photographicSensitivity int // 撮影感度(ISO感度) [TAG=0x8827]
	focalLengthIn35mmFilm   int // 35mm換算レンズ焦点距離 [TAG=0xa405]
	focalLength             int // レンズ焦点距離 [TAG=0x920a]

	dateTimeOriginal string // 原画像データの生成日時 [TAG=0x9003]
	pixelXDimension  int    // 実効画像幅 [TAG=0xa002]
	pixelYDimension  int    // 実効画像高さ [TAG=0xa003]
}

var (
	filePath = ""
)

const (
	IFD_PATH      = "IFD"
	EXIF_IFD_PATH = "IFD/Exif"
	GPS_IFD_PATH  = "IFD/GPSInfo"
)

func main() {
	flag.StringVar(&filePath, "f", "", "Path to the image file")
	flag.Parse()

	if filePath == "" {
		fmt.Println("Please provide a file path using -f flag")
		os.Exit(1)
	}

	rawExif, err := exif.SearchFileAndExtractExif(filePath)
	if err != nil {
		fmt.Println("Error SearchFileAndExtractExif:", err)
		os.Exit(1)
	}

	im, _ := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		fmt.Println("Error NewIfdMappingWithStandard:", err)
		os.Exit(1)
	}

	ti := exif.NewTagIndex()

	_, index, err := exif.Collect(im, ti, rawExif)
	if err != nil {
		fmt.Println("Error Collect:", err)
		os.Exit(1)
	}

	tagName := "Model" // edit
	rootIfd := index.RootIfd

	leafIfd, err := exif.FindIfdFromRootIfd(rootIfd, "IFD/Exif") // edit
	if err != nil {
		fmt.Println("Error FindIfdFromRootIfd:", err)
		os.Exit(1)
	}

	results, err := leafIfd.FindTagWithName(tagName) // edit
	if err != nil {
		fmt.Println("Error FindTagWithName:", err)
		os.Exit(1)
	}

	if len(results) != 1 {
		os.Exit(1)
	}

	item := results[0]

	valueRaw, err := item.Value()
	if err != nil {
		fmt.Println("Error Value:", err)
		os.Exit(1)
	}

	value := valueRaw

	fmt.Printf("Tag: %s, Value: %s\n", tagName, value)
}
