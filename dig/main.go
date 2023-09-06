package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/liangwt/note/golang/demo/dig/internal"
)

func main() {
	c := internal.Init()

	err := c.Invoke(func(g *internal.UserGateway) {
		name, err := g.GetUserName("id")
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(name)
	})
	if err != nil {
		panic(err)
	}
}

func main1() {
	c := internal.Init()

	r := gin.Default()

	err := c.Invoke(func(g *internal.UserGateway) {
		r.GET("/get_username", func(c *gin.Context) {
			name, err := g.GetUserName(c.Query("id"))
			if err != nil {
				c.JSON(500, err.Error())
				return
			}

			c.JSON(200, gin.H{
				"name": name,
			})
		})
	})
	if err != nil {
		panic(err)
	}

	r.Run()
}
