# Simple Rate Limiter
Simple rate limiter for golang with burst option

## Usage
### Rate limiter without burst
```go
import (
	"fmt"
	"time"
	"github.com/rohanprasad/simple-rate-limiter"
)

func main() {
	rl := ratelimiter.NewRateLimiter(10, 1) // 10 requests per second
	for i := 0; i < 100; i++ {
		if rl.Use() {
			// do something
		} else {
			// rate limit exceeded
		}

		time.Sleep(100 * time.Millisecond)
	}
}
```

### Rate limiter with burst
```go
import (
	"fmt"
	"time"
	"github.com/rohanprasad/simple-rate-limiter"
)

func main() {
	rl := ratelimiter.NewRateLimiterWithBurst(10, 1, 5) // 10 requests per second with burst of 5
	for i := 0; i < 100; i++ {
		rl.Wait() // will wait till rate limiter allows
		// do something
	}
}
```
