package models

type Verify struct {
	Id      int
	Addr    string
	Code    string
	Created int64
}
