package model

type Keyword struct {
  Name   string  `json:"name" bson:"name"`
  Weight float64 `json:"weight" bson:"weight"`
  POS    string  `json:"pos" bson:"pos"`
}
