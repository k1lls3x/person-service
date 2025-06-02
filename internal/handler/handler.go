package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/k1lls3x/person-service/internal/entity"
	"github.com/k1lls3x/person-service/internal/service"
	"log"
)

type Handler struct {
	personService *service.PersonService
}

func NewHandler(personService *service.PersonService) *Handler {
	return &Handler{personService: personService}
}

// UpdatePerson godoc
// @Summary Обновить данные человека по id
// @Tags persons
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param person body entity.Person true "Новые данные"
// @Success 200 {object} entity.Person
// @Failure 400 {string} string "bad request"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "server error"
// @Router /api/persons/{id} [put]
func (h *Handler) UpdatePerson(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r,"id")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	var input entity.Person

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if err := h.personService.UpdatePerson(id, &input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)

}

// CreatePerson godoc
// @Summary Создать нового человека
// @Tags persons
// @Accept json
// @Produce json
// @Param person body entity.Person true "Персона"
// @Success 201 {object} entity.Person
// @Router /api/persons [post]
func (h *Handler) CreatePerson(w http.ResponseWriter, r *http.Request) {
	log.Println("CreatePerson handler called")
	var input entity.Person

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
	}
	if input.Name == "" || input.Surname == "" {
		http.Error(w, "Name and surname are required", http.StatusBadRequest)
		return
	}
	if err := h.personService.CreatePerson(&input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(input); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// DeletePerson godoc
// @Summary Удалить человека по id
// @Tags persons
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "bad request"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "server error"
// @Router /api/persons/{id} [delete]
func (h *Handler) DeletePerson(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid id", http.StatusBadRequest)
        return
    }
    if err := h.personService.DeletePersonById(id); err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

// GetPersons godoc
// @Summary Получить список людей с фильтрами и пагинацией
// @Tags persons
// @Accept json
// @Produce json
// @Param name query string false "Имя"
// @Param surname query string false "Фамилия"
// @Param gender query string false "Пол"
// @Param nationality query string false "Национальность"
// @Param minAge query int false "Мин. возраст"
// @Param maxAge query int false "Макс. возраст"
// @Param page query int false "Страница"
// @Param pageSize query int false "Размер страницы"
// @Success 200 {array} entity.Person
// @Router /api/persons [get]
func (h *Handler) GetPersons(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	filter := entity.PersonFilter{
			Name:        getStringPtr(q.Get("name")),
			Surname:     getStringPtr(q.Get("surname")),
			Gender:      getStringPtr(q.Get("gender")),
			Nationality: getStringPtr(q.Get("nationality")),
	}


	if minAgeStr := q.Get("minAge"); minAgeStr != "" {
			if minAge, err := strconv.Atoi(minAgeStr); err == nil {
					filter.MinAge = &minAge
			} else {
					http.Error(w, "minAge must be an integer", http.StatusBadRequest)
					return
			}
	}


	if maxAgeStr := q.Get("maxAge"); maxAgeStr != "" {
			if maxAge, err := strconv.Atoi(maxAgeStr); err == nil {
					filter.MaxAge = &maxAge
			} else {
					http.Error(w, "maxAge must be an integer", http.StatusBadRequest)
					return
			}
	}


	if pageStr := q.Get("page"); pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err == nil {
					filter.Page = page
			} else {
					http.Error(w, "page must be an integer", http.StatusBadRequest)
					return
			}
	}


	if pageSizeStr := q.Get("pageSize"); pageSizeStr != "" {
			if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
					filter.PageSize = pageSize
			} else {
					http.Error(w, "pageSize must be an integer", http.StatusBadRequest)
					return
			}
	}

	persons, err := h.personService.GetPersons(filter)
	if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(persons)
}

func getStringPtr(s string) *string {
	if s == "" {
			return nil
	}
	return &s
}
