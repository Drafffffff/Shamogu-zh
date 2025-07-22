//go:build sdl

package main

import (
	"log"

	"codeberg.org/anaseto/gruid"
	sdl "codeberg.org/anaseto/gruid-sdl"
)

var driver gruid.Driver
var isFullscreen bool

func initDriver(fullscreen bool) {
	isFullscreen = fullscreen
	icon, err := base64pngToRGBA(TileImgs["favicon"])
	if err != nil {
		log.Printf("decoding window icon: %v", err)
	}
	dr := sdl.NewDriver(sdl.Config{
		TileManager: &monochromeTileManager{},
		Fullscreen:  fullscreen,
		WindowTitle: "Shamogu",
		WindowIcon:  icon,
	})
	driver = dr
}

func clearCache() {
	dr := driver.(*sdl.Driver)
	dr.ClearCache()
}
