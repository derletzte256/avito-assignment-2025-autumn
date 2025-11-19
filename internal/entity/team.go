package entity

type Team struct {
	ID      int64
	Name    string
	Members []*User
}
