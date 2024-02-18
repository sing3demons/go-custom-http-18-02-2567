package main

import (
	"fmt"

	"github.com/sing3demons/go-http-service/routes"
)

func main() {
	r := routes.NewRouter()
	r.GET("/hello/{id}", func(c routes.IContext) {
		id := c.Param("id")
		c.JSON(200, "Hello, World!"+id)
	})

	r.GET("/hello", func(c routes.IContext) {

		fmt.Println(c.GetSession())
		name := c.Query("name")
		c.JSON(200, "Hello, World! "+name)
	})

	r.POST("/hello", func(c routes.IContext) {
		var data struct {
			Name string `json:"name"`
		}
		err := c.Bind(&data)
		if err != nil {
			c.JSON(400, err.Error())
			return
		}

		c.JSON(200, map[string]string{"message": "Hello, " + data.Name + "!"})
	})

	r.Start()
}
