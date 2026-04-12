package core

import (
	"context"
	"net/http"
	"time"
	"github.com/duongtuttbn/shared-resource/go-kit/log"

	"github.com/gin-gonic/gin"
)

func UseTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		timeoutCtx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		panicChan := make(chan interface{}, 1) // used to handle panics if we can't recover

		c.Request = c.Request.WithContext(timeoutCtx)
		finished := make(chan struct{}) // to indicate handler finished
		go func() {
			defer func() {
				if p := recover(); p != nil {
					log.Error(p)
					panicChan <- p
				}
			}()
			c.Next() // calls subsequent middleware(s) and handler
			finished <- struct{}{}
		}()
		select {
		case <-timeoutCtx.Done():
			c.AbortWithStatus(http.StatusRequestTimeout)
		case <-panicChan:
			c.AbortWithStatus(http.StatusInternalServerError)
		case <-finished:
			// do nothing
		}
	}
}
