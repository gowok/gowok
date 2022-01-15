package base

import (
	"github.com/gowok/gowok"
)

type Controller struct {
	Config gowok.Config
	Models gowok.Models
}

func (c *Controller) SetConfig(config gowok.Config) {
	c.Config = config
}

func (c *Controller) SetModels(models gowok.Models) {
	c.Models = models
}
