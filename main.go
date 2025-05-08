package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
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
	Make      string // カメラメーカー [TAG=0x010f]
	Model     string // カメラモデル [TAG=0x0110]
	LensMake  string // レンズメーカー [TAG=0xa433]
	LensModel string // レンズモデル [TAG=0xa434]

	ExposureTime            string // 露出時間(シャッタースピード) [TAG=0x829a]
	FNumber                 string // Fナンバー(f値) [TAG=0x829d]
	PhotographicSensitivity string // 撮影感度(ISO感度) [TAG=0x8827]
	FocalLengthIn35mmFilm   string // 35mm換算レンズ焦点距離 [TAG=0xa405]
	FocalLength             string // レンズ焦点距離 [TAG=0x920a]

	DateTimeOriginal string // 原画像データの生成日時 [TAG=0x9003]
	PixelXDimension  string // 実効画像幅 [TAG=0xa002]
	PixelYDimension  string // 実効画像高さ [TAG=0xa003]
}

var (
	filePath string

	IFD_PATH_MAP = map[string]struct {
		tagId uint16
		path  string
	}{
		"Make":                    {0x010f, IFD_PATH},
		"Model":                   {0x0110, IFD_PATH},
		"LensMake":                {0xa433, EXIF_IFD_PATH},
		"LensModel":               {0xa434, EXIF_IFD_PATH},
		"ExposureTime":            {0x829a, EXIF_IFD_PATH},
		"FNumber":                 {0x829d, EXIF_IFD_PATH},
		"PhotographicSensitivity": {0x8827, EXIF_IFD_PATH},
		"FocalLengthIn35mmFilm":   {0xa405, EXIF_IFD_PATH},
		"FocalLength":             {0x920a, EXIF_IFD_PATH},
		"DateTimeOriginal":        {0x9003, EXIF_IFD_PATH},
		"PixelXDimension":         {0xa002, EXIF_IFD_PATH},
		"PixelYDimension":         {0xa003, EXIF_IFD_PATH},
	}
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

	rootIfd := index.RootIfd

	exifData := ExifData{}

	for tagName, tagInfo := range IFD_PATH_MAP {
		tagId := tagInfo.tagId
		ifdPath := tagInfo.path

		ifd, err := exif.FindIfdFromRootIfd(rootIfd, ifdPath)
		if err != nil {
			fmt.Println("Error FindIfdFromRootIfd:", err)
			os.Exit(1)
		}

		results, err := ifd.FindTagWithId(tagId)
		if err != nil {
			fmt.Println("Error FindTagWithName:", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Printf("Tag %s not found\n", tagName)
			continue
		}

		item := results[0]

		value, _ := item.FormatFirst()
		if err != nil {
			fmt.Println("Error Value:", err)
			os.Exit(1)
		}

		switch tagName {
		case "Make":
			exifData.Make = value
		case "Model":
			exifData.Model = value
		case "LensMake":
			exifData.LensMake = value
		case "LensModel":
			exifData.LensModel = value
		case "ExposureTime":
			exifData.ExposureTime = value
		case "FNumber":
			parts := strings.Split(value, "/")
			numerator, _ := strconv.Atoi(parts[0])
			denominator, _ := strconv.Atoi(parts[1])
			f := float32(numerator) / float32(denominator)
			output := fmt.Sprintf("%.1f", f)

			exifData.FNumber = output
		case "PhotographicSensitivity":
			exifData.PhotographicSensitivity = value
		case "FocalLengthIn35mmFilm":
			exifData.FocalLengthIn35mmFilm = value
		case "FocalLength":
			exifData.FocalLength = value
		case "DateTimeOriginal":
			t, err := time.Parse("2006:01:02 15:04:05", value)
			if err != nil {
				fmt.Println("Error parsing DateTimeOriginal:", err)
				continue
			}
			output := t.Format("2006/01/02 15:04")

			exifData.DateTimeOriginal = output
		case "PixelXDimension":
			exifData.PixelXDimension = value
		case "PixelYDimension":
			exifData.PixelYDimension = value
		}
	}
}
