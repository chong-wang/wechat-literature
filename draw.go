package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"

	"github.com/golang/freetype"
)

const (
	fontSize     = 16
	fontDPI      = 72
	fontHeight   = 4  // 字高
	fontWidth    = 4  // 字宽
	roundSpacing = 8  // 四边空白距离
	lineSpacing  = 2  // 行间距
	lining       = 1  // 表格线宽
	lineMaxRune  = 15 // 一行最多max个字符
)

var (
	Green = image.NewUniform(color.RGBA{193, 255, 193, 0})
	Gray  = image.NewUniform(color.RGBA{232, 232, 232, 0})
	Smoke = image.NewUniform(color.RGBA{245, 245, 245, 0})
	White = image.White
)

func ImageBytes(img image.Image) []byte {
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 60})
	return buf.Bytes()
}

func WriteFile(file string, img image.Image) {
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		println(err.Error())
		return
	}
	defer fp.Close()

	err = jpeg.Encode(fp, img, nil)
	if err != nil {
		println(err.Error())
		return
	}
}

func ParseFont() *freetype.Context {
	font, err := freetype.ParseFont(TTF)
	if err != nil {
		println(err.Error())
		return nil
	}
	c := freetype.NewContext()
	c.SetDPI(fontDPI)
	c.SetFont(font)
	c.SetFontSize(fontSize)
	return c
}

type Records [][]string

type Cells [][]image.Rectangle

func fix(x float64) int {
	return int(x*float64(fontDPI)*(64.0/72.0)*fontSize) >> 8
}

func CalcTable(rows Records) (Cells, image.Rectangle) {
	if len(rows) == 0 {
		return nil, image.Rectangle{}
	}

	rs := make([]int, len(rows))    // 每一行宽 上下 ^v
	cs := make([]int, len(rows[0])) // 每一列长 左右 <-->
	for i, cols := range rows {
		height := 1
		for j, col := range cols {
			n := len([]rune(col))
			if n > cs[j] {
				if n > lineMaxRune {
					cs[j] = lineMaxRune
				} else {
					cs[j] = n
				}
			}

			h := 1
			for n := len([]rune(rows[i][j])); n > lineMaxRune; n -= lineMaxRune {
				h++
			}
			if h > height {
				height = h
			}
		}
		rs[i] = height
	}

	startX, startY := roundSpacing, roundSpacing
	endX, endY := 0, 0
	cells := make([][]image.Rectangle, len(rows))
	for i := 0; i < len(rs); i++ {
		cells[i] = make([]image.Rectangle, len(rows[i]))
		sx := startX + lining
		startY += lining
		dy := startY + roundSpacing + fix(float64(rs[i])*fontHeight) + (rs[i]-1)*lineSpacing + roundSpacing
		for j := 0; j < len(cs); j++ {
			dx := sx + roundSpacing + fix(fontWidth*float64(cs[j])) + roundSpacing
			cells[i][j] = image.Rect(sx, startY, dx, dy)
			sx = dx + lining

			endX = dx + lining + roundSpacing
		}
		startY = dy
		endY = startY + lining + roundSpacing
	}
	return cells, image.Rect(0, 0, endX, endY)
}

func DrawTable(img draw.Image, cells Cells) {
	l := lining
	for _, cols := range cells {
		for _, col := range cols {
			up := image.Rect(col.Min.X-l, col.Min.Y-l, col.Max.X+l, col.Min.Y)
			draw.Draw(img, up, image.Black, image.ZP, draw.Src)

			down := image.Rect(col.Min.X-l, col.Max.Y, col.Max.X+l, col.Max.Y+l)
			draw.Draw(img, down, image.Black, image.ZP, draw.Src)

			left := image.Rect(col.Min.X-l, col.Min.Y-l, col.Min.X, col.Max.Y+l)
			draw.Draw(img, left, image.Black, image.ZP, draw.Src)

			right := image.Rect(col.Max.X, col.Min.Y-l, col.Max.X+l, col.Max.Y+l)
			draw.Draw(img, right, image.Black, image.ZP, draw.Src)
		}
	}
}

func DrawRecords(img draw.Image, cells Cells, records Records, colors []image.Image) {
	c := ParseFont()
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.Black)

	for i, cols := range cells {
		for j, col := range cols {
			if i <= len(colors) && colors[i] != nil {
				draw.Draw(img, col, colors[i], image.ZP, draw.Src)
			}

			r := []rune(records[i][j])
			y := col.Min.Y + roundSpacing
			for a := 0; a < len(r); a += lineMaxRune {
				b := a + lineMaxRune
				if b > len(r) {
					b = len(r)
				}
				y += fix(fontHeight)
				c.DrawString(string(r[a:b]), freetype.Pt(col.Min.X+roundSpacing, y))
				y += lineSpacing
			}
		}
	}
}

/*
func main() {
	records := Records{{"第一行第一列", "第一行第二列"}, {"第二行第一列", "第二行第二列一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六"}, {"abc", "123"}, {"四行", "四二"}}
	cells, rect := CalcTable(records)
	fmt.Println(rect)
	img := image.NewRGBA(rect)
	draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)
	DrawTable(img, cells)
	DrawRecords(img, cells, records)
	WriteFile("image.jpg", img)
}
*/
