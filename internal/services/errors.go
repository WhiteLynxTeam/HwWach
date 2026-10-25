package services

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

// AppError ошибка с HTTP-статусом и понятным сообщением для клиента
type AppError struct {
	Status int
	Msg    string
}

func (e *AppError) Error() string { return e.Msg }

func Forbidden(msg string) error   { return &AppError{http.StatusForbidden, msg} }
func NotFound(msg string) error    { return &AppError{http.StatusNotFound, msg} }
func Conflict(msg string) error    { return &AppError{http.StatusConflict, msg} }
func BadRequest(msg string) error  { return &AppError{http.StatusBadRequest, msg} }
func Unavailable(msg string) error { return &AppError{http.StatusServiceUnavailable, msg} }

var errNoRights = Forbidden("Недостаточно прав")

// isUniqueViolation нарушение UNIQUE (в том числе частичных индексов)
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
