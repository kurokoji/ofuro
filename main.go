package main

import (
	"context"
	"fmt"
	"log"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func getFPSString(fileName string) (*string, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	data, err := ffprobe.ProbeURL(ctx, fileName)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fpsString := data.Streams[0].RFrameRate

	return &fpsString, nil
}

type Config struct {
	FileName       string
	OutputFileName string
	FlowTime       int
	FontSize       int
	FpsString      string
	LinePadding    int
}

type FlowTimeRange struct {
	StartTime int
	EndTime   int
}

type FlowPatcher struct {
	Conf   *Config
	Stream *ffmpeg.Stream
	Lines  [5]*FlowTimeRange
}

func (f *FlowPatcher) FlowText(text string, startTime int) *FlowPatcher {
	lineNum := 0
	isOverlap := true

	linesLen := len(f.Lines)
	for idx := 0; idx < linesLen; idx++ {
		line := f.Lines[idx]
		if line == nil || !(line.StartTime <= startTime && startTime <= line.EndTime-f.Conf.FlowTime/2) {
			f.Lines[idx] = &FlowTimeRange{
				StartTime: startTime,
				EndTime:   startTime + f.Conf.FlowTime,
			}
			lineNum = idx
			isOverlap = false
			break
		}
	}

	if isOverlap {
		f.Lines[0] = &FlowTimeRange{
			StartTime: startTime,
			EndTime:   startTime + f.Conf.FlowTime,
		}
	}

	expXString := fmt.Sprintf("w-(w+tw)*((n-%d*%s)/(%s*%d))", startTime, f.Conf.FpsString, f.Conf.FpsString, f.Conf.FlowTime)

	res := f.Stream.Drawtext(text, 0, f.Conf.LinePadding*lineNum+f.Conf.FontSize*lineNum, false, ffmpeg.KwArgs{
		"enable":     fmt.Sprintf("between(t,%d,%d)", startTime, startTime+f.Conf.FlowTime),
		"fontcolor":  "white",
		"fontsize":   f.Conf.FontSize,
		"fontfile":   "/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc",
		"borderw":    2,
		"box":        0,
		"boxborderw": 10,
		"alpha":      0.8,
		"x":          expXString,
	})

	return &FlowPatcher{f.Conf, res, f.Lines}
}

func (f *FlowPatcher) Run() error {
	return f.Stream.Output(f.Conf.OutputFileName).OverWriteOutput().Run()
}

func NewFlowPatcher(fileName string, outputFileName string, flowTime int) (*FlowPatcher, error) {
	fpsString, err := getFPSString(fileName)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	conf := &Config{
		FileName:       fileName,
		OutputFileName: outputFileName,
		FlowTime:       flowTime,
		FontSize:       50,
		FpsString:      *fpsString,
		LinePadding:    15,
	}

	input := ffmpeg.Input(fileName)

	var lines [5]*FlowTimeRange

	return &FlowPatcher{
		Conf:   conf,
		Stream: input,
		Lines:  lines,
	}, nil
}

func main() {
	patcher, err := NewFlowPatcher("./peco.mp4", "./peco-text.mp4", 5)
	if err != nil {
		log.Println(err)
		return
	}

	err = patcher.FlowText("ほげ", 0).FlowText("ほげほげほげほげほげひおげほげ", 4).FlowText("popopo", 4).FlowText("popopopopo", 10).Run()

	if err != nil {
		log.Println(err)
		return
	}
}
