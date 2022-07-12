package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

type Form interface {
	message() map[string]map[string]string
}

func ParseError(form Form, ve validator.ValidationErrors) gin.H {
	m := gin.H{}
	message := form.message()
	for _, e := range ve {
		if str, ok := message[e.Field()][e.Tag()]; ok {
			m[e.Field()] = str
		}
	}
	return m
}
