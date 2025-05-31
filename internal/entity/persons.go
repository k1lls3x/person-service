package entity

type Person struct {
	ID          int     `db:"id" json:"id"`
	Name        string  `db:"name" json:"name" validate:"required"`
	Surname     string  `db:"surname" json:"surname" validate:"required"`
	Patronymic  *string `db:"patronymic" json:"patronymic,omitempty"`
	Age         *int    `db:"age" json:"age,omitempty"`
	Gender      *string `db:"gender" json:"gender,omitempty"`
	Nationality *string `db:"nationality" json:"nationality,omitempty"`
	CreatedAt   string  `db:"created_at" json:"created_at"`
}
