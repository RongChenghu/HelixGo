package resp

import "github.com/gin-gonic/gin"

func JSONOK(c *gin.Context, payload interface{}) {
	c.JSON(200, payload)
}

func JSONError(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{
		"error":   code,
		"message": message,
	})
}
