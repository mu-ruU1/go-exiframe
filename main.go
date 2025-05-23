package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/math/fixed"
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
	PixelXDimension  int    // 実効画像幅 [TAG=0xa002]
	PixelYDimension  int    // 実効画像高さ [TAG=0xa003]
	Orientation      string // 画像の向き [TAG=0x0112]
}

// go-exiframeの設定
type Config struct {
	filePath        string
	frameColorBlack bool
	noFrame         bool
	noModelData     bool

	fileName   string
	frameColor *image.Uniform
	textColor  *image.Uniform
}

var (
	filePath     string
	framePixel   int = 180
	noFramePixel int = 0

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
		"Orientation":             {0x0112, IFD_PATH},
	}
)

const (
	IFD_PATH      = "IFD"
	EXIF_IFD_PATH = "IFD/Exif"
	GPS_IFD_PATH  = "IFD/GPSInfo"

	EXIF_LABEL_HEIGHT = 600
	LARGE_FONT_SIZE   = 200
	FONT_SIZE         = 150

	FILE_NAME_PREFIX = "exiframe-"
)

func getExif(config *Config) (exifData *ExifData) {
	rawExif, err := exif.SearchFileAndExtractExif(config.filePath)
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

	exifData = &ExifData{}

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
			var output string
			if strings.Contains(value, "/") {
				parts := strings.Split(value, "/")
				numerator, _ := strconv.Atoi(parts[0])
				denominator, _ := strconv.Atoi(parts[1])
				f := float32(numerator) / float32(denominator)
				output = fmt.Sprintf("%.1f", f)
			} else {
				output = value
			}

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
			output, err := strconv.Atoi(value)
			if err != nil {
				fmt.Println("Error parsing PixelXDimension:", err)
				os.Exit(1)
			}

			exifData.PixelXDimension = output
		case "PixelYDimension":
			output, err := strconv.Atoi(value)
			if err != nil {
				fmt.Println("Error parsing PixelYDimension:", err)
				os.Exit(1)
			}

			exifData.PixelYDimension = output
		case "Orientation":
			exifData.Orientation = value
		}
	}

	return exifData
}

func drawFrame(config *Config, exifData *ExifData) {
	fSrc, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer fSrc.Close()

	src, err := imaging.Open(filePath, imaging.AutoOrientation(true))
	if err != nil {
		fmt.Println("Error Decode:", err)
		os.Exit(1)
	}

	// 画像のサイズを取得
	srcBounds := src.Bounds()
	srcWidth := srcBounds.Max.X
	srcHeight := srcBounds.Max.Y

	if config.noFrame {
		framePixel = 0
		noFramePixel = 180
	}

	// 背景フレームの作成
	dst := image.NewRGBA(image.Rect(0, 0, srcWidth+framePixel*2, srcHeight+framePixel*2+EXIF_LABEL_HEIGHT+noFramePixel))

	if config.frameColorBlack {
		config.frameColor = image.Black
		config.textColor = image.White
	} else {
		config.frameColor = image.White
		config.textColor = image.Black
	}

	draw.Draw(dst, dst.Bounds(), config.frameColor, image.Point{}, draw.Src)

	// 画像と背景フレームの描画
	draw.Draw(dst, dst.Bounds(), src, image.Point{-framePixel, -framePixel}, draw.Src)

	// exiframe-*.jpg として保存
	fDst, err := os.Create(FILE_NAME_PREFIX + config.fileName)
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
	defer fDst.Close()

	// Exif情報をJPEGに埋め込む
	var camData, lensData string
	if !config.noModelData {
		camData = exifData.Make + " " + exifData.Model
		lensData = exifData.LensMake + " " + exifData.LensModel
	}

	expoData := exifData.FocalLengthIn35mmFilm + "mm  " + "f/" + exifData.FNumber + "  " + exifData.ExposureTime + "s  ISO" + exifData.PhotographicSensitivity

	boldfnt, err := truetype.Parse(gomonobold.TTF)
	if err != nil {
		fmt.Println("Error parsing font:", err)
		os.Exit(1)
	}

	regularfnt, err := truetype.Parse(gomono.TTF)
	if err != nil {
		fmt.Println("Error parsing font:", err)
		os.Exit(1)
	}

	boldFace := truetype.NewFace(boldfnt, &truetype.Options{
		Size: LARGE_FONT_SIZE,
	})

	boldFace2 := truetype.NewFace(boldfnt, &truetype.Options{
		Size: FONT_SIZE,
	})

	regularFace := truetype.NewFace(regularfnt, &truetype.Options{
		Size: FONT_SIZE,
	})

	boldMetrics, regularMetrics := boldFace.Metrics(), regularFace.Metrics()
	boldHeight, _ := boldMetrics.Height.Ceil(), regularMetrics.Height.Ceil()

	dBold := &font.Drawer{
		Dst:  dst,
		Src:  config.textColor,
		Face: boldFace,
		Dot:  fixed.Point26_6{},
	}

	dBold2 := &font.Drawer{
		Dst:  dst,
		Src:  config.textColor,
		Face: boldFace2,
		Dot:  fixed.Point26_6{},
	}

	dRegular := &font.Drawer{
		Dst:  dst,
		Src:  config.textColor,
		Face: regularFace,
		Dot:  fixed.Point26_6{},
	}

	// カメラデータ
	dBold.Dot.X = fixed.I(framePixel + noFramePixel)
	dBold.Dot.Y = fixed.I(srcHeight + framePixel*2 + boldHeight + noFramePixel)
	dBold.DrawString(camData)

	// レンズデータ
	dRegular.Dot.X = fixed.I(framePixel + noFramePixel)
	dRegular.Dot.Y = fixed.I(srcHeight + framePixel*2 + boldHeight*2 + noFramePixel)
	dRegular.DrawString(lensData)

	// 撮影データ
	expoDataWidth := dBold2.MeasureString(expoData).Ceil()
	dBold2.Dot.X = fixed.I(srcWidth + framePixel - expoDataWidth - noFramePixel)
	dBold2.Dot.Y = fixed.I(srcHeight + framePixel*2 + boldHeight + noFramePixel)
	dBold2.DrawString(expoData)

	// 撮影日時
	timeDataWidth := dRegular.MeasureString(exifData.DateTimeOriginal).Ceil()
	dRegular.Dot.X = fixed.I(srcWidth + framePixel - timeDataWidth - noFramePixel)
	dRegular.Dot.Y = fixed.I(srcHeight + framePixel*2 + boldHeight*2 + noFramePixel)
	dRegular.DrawString(exifData.DateTimeOriginal)

	// JPEGエンコード
	err = jpeg.Encode(fDst, dst, &jpeg.Options{Quality: 100})
	if err != nil {
		fmt.Println("Error encoding JPEG:", err)
		os.Exit(1)
	}
}

func main() {
	flag.StringVar(&filePath, "f", "", "Path to the image file (required)")
	frameColorBlack := flag.Bool("black", false, "Use black color frame (default white)")
	noFrame := flag.Bool("no-frame", false, "Do not draw frame (default draw frame)")
	noModelData := flag.Bool("no-model", false, "Do not draw model data (default draw model data)")
	flag.Parse()

	if filePath == "" {
		fmt.Println("Please provide a file path using -f flag")
		os.Exit(1)
	}

	fileName := filepath.Base(filePath)

	config := &Config{
		filePath:        filePath,
		frameColorBlack: *frameColorBlack,
		noFrame:         *noFrame,
		noModelData:     *noModelData,

		fileName: fileName,
	}

	exifData := getExif(config)

	drawFrame(config, exifData)
}
