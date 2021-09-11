package main

import (
	"github.com/aokabi/ngraphinx/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}

// func main() {
// 	// gin server
// 	r := gin.Default()
// 	r.GET("/graph", func(c *gin.Context) {
// 		param, ok := c.GetQuery("aggregates")
// 		aggregates := []string{}
// 		if ok {
// 			aggregates = strings.Split(param, ",")
// 		}
// 		err := lib.GenerateGraph(aggregates)
// 		if err != nil {
// 			c.JSON(500, gin.H{
// 				"message": err.Error(),
// 			})
// 			return
// 		}
// 	})
// 	r.Run("127.0.0.1:8889")
// }
