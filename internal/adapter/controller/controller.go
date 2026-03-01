package controller

import (
	"github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/controller/middleware"
	v1 "github.com/SosisterRapStar/LETI-PaperTestMicroservices/internal/adapter/controller/v1"
)

type Controller struct {
	Middleware middleware.Middleware
	V1         v1.Controller
}
