## Circuit Breaker tool
Implementation of the Circuit Breaker pattern. Implemented using the `sync/atomic` package.

![Circuit Breaker scheme](https://github.com/leonidkit/circuitbreaker/blob/master/circuitbreaker.png?raw=true)

### Features:
* Use `Interval` to set the interval after which it will clear counters. Default store all counters of requests (successful, throtlled, etc.).
* Use `Timeout` to set the interval after which it will switch to the HalfClosed state. Default 1 second.
* Use `Treshold` to set threshold value for consecutive errors. Default 1.
* Use `MaxRequests` to set value for max request number in state HalfClosed.

### Example
```
cb := circuitbreaker.New(circuitbreaker.Settings{
    Timeout:     2 * time.Second,
    Threshold:   2,
    MaxRequests: 2,
})

for i := 0; i < 10; i++ {
    if !cb.Allow() {
        continue
    }
    if i % 2 == 0 {
        cb.RegisterError()
    } else {
        cb.RegisterOK()
    }
}
```