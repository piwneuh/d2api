package models

type Player struct {
	Id      uint32   `bson:"_id" json:"id"`
	Matches []string `bson:"matches" json:"matches"`
}
