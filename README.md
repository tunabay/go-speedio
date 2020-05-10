# go-speedio

[![GitHub](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/tunabay/go-speedio/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/tunabay/go-speedio?status.svg)](https://godoc.org/github.com/tunabay/go-speedio)

speedio is a Go package implementing bit rate limiting and bit rate
measurement. It wraps an `io.Reader` or `io.Writer` object.

## Usage

```
import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/tunabay/go-speedio"
)

func main() {
	// /dev/null writer with rate limit of 256 bit/s
	w, err := speedio.NewWriter(ioutil.Discard, 256)
	if err != nil {
		panic(err)
	}

	// print bit rate every 0.5 seconds
	go func() {
		for range time.Tick(time.Second / 2) {
			fmt.Println("bitrate:", w.BitRate())
		}
	}()

	// start measurement
	w.Start()
	time.Sleep(time.Second) // measurement resolution

	// write 160 bytes data, it will take about 5 seconds
	buf := make([]byte, 160)
	n, err := w.Write(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println("written:", n)

	// close wrapper only (don't want to close /dev/null)
	if err := w.CloseSingle(); err != nil {
		panic(err)
	}

	// print total
	bc, et, br := w.Total()
	fmt.Println("total byte count:", bc)
	fmt.Println("elapsed time:", et)
	fmt.Println("total bit rate:", br)
}
```
[Run in Go Playground](https://play.golang.org/p/mAIH6Jh5_kF)

## Documentation

- http://godoc.org/github.com/tunabay/go-speedio

## License

go-speedio is available under the MIT license. See the [LICENSE](LICENSE) file
for more information.
