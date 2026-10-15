//go:build sdl

package main

import (
	"log"

	"codeberg.org/anaseto/gruid"
	sdl "codeberg.org/anaseto/gruid-sdl"
)

var driver gruid.Driver

func initDriver(fullscreen bool, sx, sy float64) {
	icon, err := base64pngToRGBA(TileImgs["favicon"])
	if err != nil {
		log.Printf("解码窗口图标失败：%v", err)
	}
	dr := sdl.NewDriver(sdl.Config{
		TileManager: &monochromeTileManager{},
		Fullscreen:  fullscreen,
		WindowTitle: "萨满山灵守护者 · Shamogu",
		WindowIcon:  icon,
	})
	if sx != 1 || sy != 1 {
		dr.SetScale(float32(sx), float32(sy))
	}
	driver = dr
}

func clearCache() {
	dr := driver.(*sdl.Driver)
	dr.ClearCache()
}
