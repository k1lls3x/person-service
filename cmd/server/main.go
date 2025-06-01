package main

import (
	_"context"
	"log"
	_"time"

	"github.com/k1lls3x/person-service/internal/entity"
	"github.com/k1lls3x/person-service/internal/repository"
	"github.com/k1lls3x/person-service/internal/service"
)

func main(){
	// Контекст с таймаутом
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	repository.Init()
	// Тестовый пользователь
	p := &entity.Person{
		Name:    "Kiska",
		Surname: "Solevaya",
	}

	// Вызов CreatePerson
	if err := service.CreatePerson( p); err != nil {
		log.Printf("failed to create person: %v", err)
	} else {
		log.Printf("person created successfully: %+v", p)
	}
}
