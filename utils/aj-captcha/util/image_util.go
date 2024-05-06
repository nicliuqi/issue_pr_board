package util

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

type ImageUtil struct {
	Src       string
	SrcImage  image.Image
	RgbaImage *image.RGBA
	FontPath  string
	Width     int
	Height    int
}

func NewImageUtil(src string, fontPath string) *ImageUtil {
	srcImage := OpenPngImage(src)

	return &ImageUtil{Src: src,
		SrcImage:  srcImage,
		RgbaImage: ImageToRGBA(srcImage),
		Width:     srcImage.Bounds().Dx(),
		Height:    srcImage.Bounds().Dy(),
		FontPath:  fontPath,
	}
}

func (i *ImageUtil) IsOpacity(x, y int) bool {
	A := i.RgbaImage.RGBAAt(x, y).A

	if float32(A) <= 125 {
		return true
	}
	return false
}

func (i *ImageUtil) DecodeImageToFile() {
	filename := "drawImg.png"
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("创建 %s 失败 %v", filename, err)
	}

	err = png.Encode(file, i.RgbaImage)
	if err != nil {
		log.Printf("png %s Encode 失败 %v", filename, err)
	}

}

func (i *ImageUtil) SetPixel(rgba color.RGBA, x, y int) {
	i.RgbaImage.SetRGBA(x, y, rgba)
}

func (i *ImageUtil) Base64() (string, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, i.RgbaImage)

	if err != nil {
		log.Printf("img写入buf失败 err: %v", err)
		return "", err
	}

	dist := make([]byte, buf.Cap()+buf.Len())
	base64.StdEncoding.Encode(dist, buf.Bytes())
	return string(dist), nil
}

func (i *ImageUtil) VagueImage(x int, y int) {
	var red uint32
	var green uint32
	var blue uint32
	var alpha uint32

	points := [8][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}

	for _, point := range points {
		pointX := x + point[0]
		pointY := y + point[1]

		if pointX < 0 || pointX >= i.Width || pointY < 0 || pointY >= i.Height {
			continue
		}

		r, g, b, a := i.RgbaImage.RGBAAt(pointX, pointY).RGBA()
		red += r >> 8
		green += g >> 8
		blue += b >> 8
		alpha += a >> 8

	}

	var avg uint32
	avg = 8

	rgba := color.RGBA{R: uint8(red / avg), G: uint8(green / avg), B: uint8(blue / avg), A: uint8(alpha / avg)}

	i.RgbaImage.SetRGBA(x, y, rgba)
}

func OpenPngImage(src string) image.Image {
	ff, err := os.Open(src)
	if err != nil {
		log.Printf("打开 %s 图片失败: %v", src, err)
	}

	img, err := png.Decode(ff)

	if err != nil {
		log.Printf("png %s decode  失败: %v", src, err)
	}

	return img
}

func ImageToRGBA(img image.Image) *image.RGBA {
	if dst, ok := img.(*image.RGBA); ok {
		return dst
	}

	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	return dst
}
