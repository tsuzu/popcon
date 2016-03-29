package main

// User is a struct to save UserData
type User struct {
	internalID int
	userID     string
	userName   string
	passHash   [64]byte
	email      string
}
