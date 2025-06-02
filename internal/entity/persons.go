package entity

import "time"

type Person struct {
	ID          int     `db:"id" json:"id"`
	Name        string  `db:"name" json:"name" validate:"required"`
	Surname     string  `db:"surname" json:"surname" validate:"required"`
	Patronymic  *string `db:"patronymic" json:"patronymic,omitempty"`
	Age         *int    `db:"age" json:"age,omitempty"`
	Gender      *string `db:"gender" json:"gender,omitempty"`
	Nationality *string `db:"nationality" json:"nationality,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   string  `db:"updated_at" json:"updated_at"`
}


type PersonFilter struct {
	Name        *string `form:"name" json:"name,omitempty"`
	Surname     *string `form:"surname" json:"surname,omitempty"`
	Gender      *string `form:"gender" json:"gender,omitempty"`
	Nationality *string `form:"nationality" json:"nationality,omitempty"`
	MinAge      *int    `form:"min_age" json:"min_age,omitempty"`
	MaxAge      *int    `form:"max_age" json:"max_age,omitempty"`
	Page        int     `form:"page" json:"page" validate:"gte=1"`          // по умолчанию 1
	PageSize    int     `form:"page_size" json:"page_size" validate:"gte=1,lte=100"` // ограничение на размер страницы
}
