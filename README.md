# go-exiframe

| `go-exiframe -f ./sample01.jpeg -black` | `go-exiframe -f ./sample02.jpeg -no-frame` |
| :-------------------------------------: | :----------------------------------------: |
|  ![](./sample/exiframe-sample01.jpeg)   |    ![](./sample/exiframe-sample02.jpeg)    |

## 使い方

```bash
# Install
$ go install github.com/mu-ruU1/go-exiframe@latest

# Help
$ go-exiframe -h
Usage of go-exiframe:
  -black
        Use black color frame (default white)
  -f string
        Path to the image file (required)
  -no-frame
        Do not draw frame (default draw frame)
  -no-model
        Do not draw model data (default draw model data)

# Example
$ go-exiframe -f /path/to/image.jpg
## Export file to exiframe-image.jpg
```

## 参考

- [go-exif/v3](https://pkg.go.dev/github.com/dsoprea/go-exif/v3)
- [CIPA DC-008-2024](https://www.cipa.jp/j/std/std-sec.html)
  - デジタルスチルカメラ用画像ファイルフォーマット規格 Exif 3.0
