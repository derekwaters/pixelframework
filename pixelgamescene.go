package pixelframework

import (
	"fmt"

	"github.com/gopxl/pixel/v2/backends/opengl"
)

type PixelGameScene interface {
	init(win opengl.Window) error
	render() error
	destroy()
}

func Hello(name string) string {
	message := fmt.Sprintf("Hi, %v. Welcome!", name)
	return message
}