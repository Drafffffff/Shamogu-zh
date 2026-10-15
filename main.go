//go:build !js

package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/ui"
)

func main() {
	optReplay := flag.String("r", "", "回放文件路径（_ 表示默认位置）")
	optGameLogs := flag.Bool("l", false, "将游戏日志写入文件")
	optVersion := flag.Bool("version", false, "显示构建信息")
	optOptions := flag.String("O", "", "选项（说明见 vars.go 文件）")
	opt16colors, opt256colors, optTrueColor := new(bool), new(bool), new(bool)
	optFullscreen := new(bool)
	optWidthScale, optHeightScale := new(float64), new(float64)
	if Tiles {
		optFullscreen = flag.Bool("F", false, "全屏")
		optWidthScale = flag.Float64("w", 1.0, "窗口宽度缩放比例（默认：1.0，示例：0.75、1.25）")
		optHeightScale = flag.Float64("h", 1.0, "窗口高度缩放比例（默认：1.0，示例：0.75、1.25）")
	} else {
		opt16colors = flag.Bool("s", false, "使用标准 16 色调色板（多数系统默认）")
		opt256colors = flag.Bool("x", false, "使用 xterm 256 色调色板（近似 Solarized 配色）")
		optTrueColor = flag.Bool("t", false, "使用真彩色 Selenized 调色板（并非所有终端都支持）")
	}
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "用法：shamogu [选项]\n\n可用选项：")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *optVersion {
		fmt.Printf("shamogu\t%v\n", Version)
		if bi, ok := debug.ReadBuildInfo(); ok {
			fmt.Print(bi)
		}
		os.Exit(0)
	}
	processOptions(*optOptions)
	if runtime.GOOS == "windows" {
		ColorMode = ColorMode8
	}
	switch {
	case *opt256colors:
		ColorMode = ColorMode256
	case *opt16colors:
		ColorMode = ColorMode16
	case *optTrueColor:
		ColorMode = ColorMode24bit
	}
	log.SetPrefix("shamogu ")
	err := InitConfig()
	if err != nil {
		log.Print(err)
	}
	initDriver(*optFullscreen, *optWidthScale, *optHeightScale)
	if *optReplay != "" {
		RunReplay(*optReplay)
	} else {
		RunGame(*optGameLogs)
	}
}

// RunGame starts the game using the given path for the log file (if
// non-empty).
func RunGame(gamelogs bool) {
	gd := gruid.NewGrid(UIWidth, UIHeight)
	md := &model{gd: gd, g: &Game{}, targ: &targeting{}}
	var repw io.WriteCloser
	dir, err := DataDir()
	defer func() {
		if repw != nil {
			if err := repw.Close(); err != nil {
				log.Printf("关闭回放文件失败：%v", err)
			}
		}
		if md.gameEnded && dir != "" {
			if err := RemoveSaveFile(); err != nil {
				log.Printf("删除存档文件失败：%v", err)
			}
			_, err := os.Stat(filepath.Join(dir, "replay.part"))
			if err != nil {
				log.Printf("未找到回放文件：%v", err)
				return
			}
			if err := os.Rename(filepath.Join(dir, "replay.part"), filepath.Join(dir, "replay")); err != nil {
				log.Printf("写入回放文件失败：%v", err)
			}
		}
	}()
	if err == nil {
		replay, err := os.OpenFile(filepath.Join(dir, "replay.part"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err == nil {
			repw = replay
		} else {
			log.Printf("写入回放数据失败：%v", err)
		}
	} else {
		log.Print(err)
	}
	if gamelogs {
		LogGame = true
	}
	app := gruid.NewApp(gruid.AppConfig{
		Driver:      driver,
		Model:       md,
		FrameWriter: repw,
	})
	if f := setLogOutput(); f != nil {
		defer func() {
			f.Close()
		}()
	}
	err = app.Start(context.Background())
	log.SetOutput(os.Stderr)
	if err != nil {
		log.Fatal(err)
	}
}

// RunReplay runs the given replay file.
func RunReplay(file string) {
	if file == "_" {
		dir, err := DataDir()
		if err == nil {
			file = filepath.Join(dir, "replay")
		} else {
			log.Print(err)
		}
	}
	replay, err := os.Open(file)
	if err != nil {
		log.Fatalf("载入回放文件失败：%v", err)
	}
	defer replay.Close()
	fd, err := gruid.NewFrameDecoder(replay)
	if err != nil {
		// Attempt again assuming base64-encoded replay (from web version).
		replay.Seek(0, io.SeekStart)
		b64replay := base64.NewDecoder(base64.StdEncoding, replay)
		var errb64 error
		fd, errb64 = gruid.NewFrameDecoder(b64replay)
		if errb64 != nil {
			log.Fatalf("帧解码器错误：%v", err)
		}
	}
	gd := gruid.NewGrid(UIWidth, UIHeight)
	rep := ui.NewReplay(ui.ReplayConfig{
		Grid:         gd,
		FrameDecoder: fd,
	})
	app := gruid.NewApp(gruid.AppConfig{
		Driver: driver,
		Model:  rep,
	})
	if f := setLogOutput(); f != nil {
		defer func() {
			f.Close()
		}()
	}
	if err := app.Start(context.Background()); err != nil {
		log.SetOutput(os.Stderr)
		log.Fatal(err)
	}
}

// setLogOutput sets standard log output to the logs file in the game's data
// directory.
func setLogOutput() *os.File {
	dataDir, err := DataDir()
	if err != nil {
		log.Print(err)
		return nil
	}
	logFile := filepath.Join(dataDir, "logs.txt")
	f, err := os.Create(logFile)
	if err != nil {
		log.Print(err)
		return nil
	}
	if Tiles {
		log.SetOutput(io.MultiWriter(f, os.Stderr))
	} else {
		log.SetOutput(f)
	}
	return f
}

// subSig is a subscription that intercepts SIGTERM for closing the game
// gracefully.
func subSig(ctx context.Context, msgs chan<- gruid.Msg) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)
	select {
	case <-ctx.Done():
	case <-sig:
		msgs <- gruid.MsgQuit{}
	}
}

func processOptions(s string) {
	for opt := range strings.SplitSeq(s, ",") {
		if len(opt) == 0 {
			continue
		}
		negate := false
		if opt[0] == '!' {
			opt = opt[1:]
			negate = true
		}
		switch opt {
		case "NoChaos":
			NoChaos = !negate
			MonsData[ChaosMegabat].Name = GetMegabatName()
		case "NoAnim":
			NoAnim = !negate
		default:
			fmt.Fprintf(os.Stderr, "shamogu：未知选项：%v\n", opt)
		}
	}
}
