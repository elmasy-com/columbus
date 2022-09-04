package closer

import (
	"fmt"

	"github.com/elmasy-com/columbus/fetcher"
	"github.com/elmasy-com/columbus/writer"
)

func Closer() {

	fmt.Printf("Closing fetcher...\n")
	fetcher.Close()

	fmt.Printf("Closing writer...\n")
	writer.Close()
}
