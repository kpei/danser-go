package storyboard

import (
	"github.com/wieku/danser-go/app/animation"
	"math"
	"strconv"
)

type LoopProcessor struct {
	start, repeats int64
	transforms     []*animation.Transformation
}

func NewLoopProcessor(data []string) *LoopProcessor {
	loop := new(LoopProcessor)
	loop.start, _ = strconv.ParseInt(data[1], 10, 64)
	loop.repeats, _ = strconv.ParseInt(data[2], 10, 64)
	return loop
}

func (loop *LoopProcessor) Add(command *Command) {
	loop.transforms = append(loop.transforms, command.GenerateTransformations()...)
}

func (loop *LoopProcessor) Unwind() []*animation.Transformation {
	var transforms []*animation.Transformation

	startTime := math.MaxFloat64
	endTime := -math.MaxFloat64

	for _, t := range loop.transforms {
		startTime = math.Min(startTime, t.GetStartTime())
		endTime = math.Max(endTime, t.GetEndTime())
	}

	transTime := endTime - startTime

	for i := int64(0); i < loop.repeats; i++ {
		partStart := float64(loop.start) + float64(i)*transTime
		if i > 0 {
			partStart -= startTime
		}

		for _, t := range loop.transforms {
			transforms = append(transforms, t.Clone(partStart+t.GetStartTime(), partStart+t.GetEndTime()))
		}
	}

	return transforms
}
