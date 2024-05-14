package models

type Player struct {
	Id      int      `bson:"_id" json:"id"`
	Matches []uint64 `bson:"" json:"matches"`
}
