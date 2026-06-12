package httpserver

import (
	creativerunsvc "github.com/woragis/ecom-op-creatives-backend/server/internal/creativerun/service"
	productsvc "github.com/woragis/ecom-op-creatives-backend/server/internal/product/service"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/platform/rabbitmq"
	"gorm.io/gorm"
)

type App struct {
	DB       *gorm.DB
	RabbitMQ *rabbitmq.Client
	Products *productsvc.Service
	Runs     *creativerunsvc.Service
}
