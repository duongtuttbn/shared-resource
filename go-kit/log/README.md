# log

Init before first use

```golang
package main

import (
	"git.xantus.network/shared-resources/go-kit/log"
)

func main()  {
	log.SetDefault(log.Config{
		Level:  "debug",
		Format: "text",
	})
}
```
