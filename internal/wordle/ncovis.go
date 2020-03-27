package wordle

import (
	"context"
	"time"

	"github.com/vivym/chinanews-spider/internal/api/protobuf/wordle"
	"github.com/vivym/chinanews-spider/internal/model"
)

func (n *WordleToolkit) NCOVIS_ShapeWordle(words []model.Keyword, region string) ([]model.Word, []model.Word, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	var ws []*wordle.NCOVIS_ShapeWordleRequest_Word
	for _, k := range words {
		ws = append(ws, &wordle.NCOVIS_ShapeWordleRequest_Word{
			Name:   k.Name,
			Weight: k.Weight,
		})
	}
	request := wordle.NCOVIS_ShapeWordleRequest{
		Method: wordle.NCOVIS_ShapeWordleRequest_ShapeWordle,
		Topic:  "Chinanews-" + regionToShort(region),
		Words:  ws,
	}

	rsp, err := n.client.NCOVIS_ShapeWordle(ctx, &request)
	if err != nil {
		return nil, nil, err
	}

	keywords := make([]model.Word, 0, len(rsp.GetKeywords()))
	for _, w := range rsp.GetKeywords() {
		keywords = append(keywords, model.Word{
			Name:     w.Name,
			FontSize: w.FontSize,
			Color:    w.Color,
			Rotate:   w.Rotate,
			TransX:   w.TransX,
			TransY:   w.TransY,
			FillX:    w.FillX,
			FillY:    w.FillY,
		})
	}

	fillingWords := make([]model.Word, 0, len(rsp.GetFillingWords()))
	for _, w := range rsp.GetFillingWords() {
		fillingWords = append(fillingWords, model.Word{
			Name:     w.Name,
			FontSize: w.FontSize,
			Color:    w.Color,
			Rotate:   w.Rotate,
			TransX:   w.TransX,
			TransY:   w.TransY,
			FillX:    w.FillX,
			FillY:    w.FillY,
		})
	}

	return keywords, fillingWords, nil
}
