package models

type Team struct {
	ID       int    `db:"id"`
	TeamName string `db:"team_name"`

	Members []User `db:"-"`
}
