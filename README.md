# Simplepush
Library for [Simplepush](https://simplepush.io)

# Example
```
go get github.com/simplepush/simplepush-go
```

```
package main

import "github.com/simplepush/simplepush-go"

func main() {
  // Send notification
  simplepush.Send(simplepush.Message{"HuxgBB", "title", "message", "event", false, "", ""})
  // Send encrypted notification
  simplepush.Send(simplepush.Message{"HuxgBB", "title", "message", "event", true, "password", "salt"})
}
```
