package main

import (
	"image"
	"image/draw"
)

const DefaultSharedBookFile = "shared.json"

type SharedBook struct {
	Book     string `json:"book"`
	Offer    string `json:"offer"`
	Borrower string `json:"borrower"`
}

var AllSharedBooks []SharedBook

func LoadAllSharedBooks(file string) {
	LoadFromFile(&AllSharedBooks, file)
}

func sharedBooksRows() ([][]string, []image.Image) {
	rows := make([][]string, 1+len(AllSharedBooks))
	colors := make([]image.Image, len(rows))
	rows[0] = append(rows[0], "书籍", "提供者", "借阅人")
	colors[0] = Gray

	for i, book := range AllSharedBooks {
		rows[i+1] = append(rows[i+1], book.Book)
		rows[i+1] = append(rows[i+1], book.Offer)
		rows[i+1] = append(rows[i+1], book.Borrower)
	}
	return rows, colors
}

func SharedBooksImage() []byte {
	records, colors := sharedBooksRows()
	cells, rect := CalcTable(records)
	img := image.NewRGBA(rect)
	draw.Draw(img, img.Bounds(), image.White, image.ZP, draw.Src)
	DrawTable(img, cells)
	DrawRecords(img, cells, records, colors)
	return ImageBytes(img)
}
