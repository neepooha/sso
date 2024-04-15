package models

type User struct {
	ID       uint64
	Email    string
	PassHash []byte
}
