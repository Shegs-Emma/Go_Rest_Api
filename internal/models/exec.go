package models

type Exec struct {
	ID 			int		`json:"id,omitempty"`
	FirstName 	string	`json:"first_name,omitempty"`
	LastName 	string	`json:"last_name,omitempty"`
	Class 		string	`json:"class,omitempty"`
	Email		string	`json:"email,omitempty"`
	Subject 	string	`json:"subject,omitempty"`
}