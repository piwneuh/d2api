package models

type Player struct {
	Id      uint32   `bson:"_id" json:"id"`
	Matches []uint64 `bson:"matches" json:"matches"`
}
