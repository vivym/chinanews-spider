package model

import (
	"time"

	"github.com/Kamva/mgm/v2"
)

type NewsMetadata struct {
	mgm.DefaultModel `bson:",inline"`
	Province         string    `json:"province" bson:"province"`
	URL              string    `json:"url" bson:"url"`
	Tag              string    `json:"tag" bson:"tag"`
	Title            string    `json:"title" bson:"title"`
	Date             time.Time `json:"date" bson:"date"`
}

type News struct {
	mgm.DefaultModel `bson:",inline"`
	Region           string    `json:"region" bson:"region"`
	Date             time.Time `json:"date" bson:"date"`
	Tags             []Tag     `json:"tags" bson:"tags"`
	Keywords         []Word    `json:"keywords" bson:"keywords"`
	FillingWords     []Word    `json:"fillingWords" bson:"fillingWords"`
}

type Tag struct {
	Name  string `json:"name" bson:"name"`
	Count int    `json:"count" bson:"count"`
}

type Word struct {
	Name     string  `json:"name" bson:"name"`
	FontSize float64 `json:"fontSize" bson:"fontSize"`
	Color    string  `json:"color" bson:"color"`
	Rotate   float64 `json:"rotate" bson:"rotate"`
	TransX   float64 `json:"transX" bson:"transX"`
	TransY   float64 `json:"transY" bson:"transY"`
	FillX    float64 `json:"fillX" bson:"fillX"`
	FillY    float64 `json:"fillY" bson:"fillY"`
}

func (n *NewsMetadata) CollectionName() string {
	return "news_metadata"
}
