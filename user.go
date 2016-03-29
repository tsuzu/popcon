package main

// User is a struct to save UserData
type User struct {
	InternalID int
	UserID     string
	UserName   string
	PassHash   [64]byte
	Email      string
}
