package play

import (
	"math/rand"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
)

type HitResults struct {
	bottom   *sprite.Manager
	top      *sprite.Manager
	lastTime float64
	diff     *difficulty.Difficulty
	color    color2.Color
	alpha    float64
}

func NewHitResults(diff *difficulty.Difficulty) *HitResults {
	// Preload all frames to avoid stalling during gameplay
	skin.GetFrames("hit0", true)

	return &HitResults{
		bottom: sprite.NewManager(),
		top:    sprite.NewManager(),
		diff:   diff,
	}
}

func (results *HitResults) AddResult(time int64, result osu.HitResult, name string, teamIndex int, position vector.Vector2d, object objects.IHitObject) {
	var tex string
	// var particle string

	switch result & osu.BaseHitsM {
	case osu.Hit300:
		tex = "hit300"
	case osu.Hit100:
		tex = "hit100"
	case osu.Hit50:
		tex = "hit50"
	case osu.Miss:
		tex = "hit0"
	}

	switch result & osu.Additions {
	case osu.KatuAddition:
		tex += "k"
	case osu.GekiAddition:
		tex += "g"
	}

	if tex == "" {
		return
	}

	// frames := skin.GetFrames(tex, true)

	// // particles := false

	// if particle != "" && len(frames) > 0 {
	// 	particleTex := skin.GetTextureSource(particle, skin.GetSourceFromTexture(frames[0]))

	// 	if particleTex != nil {
	// 		particles = true

	// 		for i := 0; i < 150; i++ {
	// 			fadeOut := 500 + 700*rand.Float64()
	// 			direction := vector.NewVec2dRad(rand.Float64()*2*math.Pi, rand.Float64()*35)

	// 			sp := sprite.NewSpriteSingle(particleTex, float64(time)+0.5, position, vector.Centre)
	// 			sp.SetAdditive(true)
	// 			sp.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, float64(time), float64(time)+fadeOut, 1.0, 0.0))
	// 			sp.AddTransform(animation.NewVectorTransformV(animation.Move, easing.OutQuad, float64(time), float64(time)+fadeOut, position, position.Add(direction)))
	// 			sp.ResetValuesToTransforms()
	// 			sp.AdjustTimesToTransformations()
	// 			sp.ShowForever(false)

	// 			results.bottom.Add(sp)
	// 		}
	// 	}
	// }

	// hit := sprite.NewAnimation(frames, 1000.0/60, false, float64(time)+1, position, vector.Centre)
	fadeIn := float64(time + difficulty.ResultFadeIn)
	postEmpt := float64(time + difficulty.PostEmpt)
	fadeOut := postEmpt + float64(difficulty.ResultFadeOut) + 350

	// hit.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(time), fadeIn, 0.0, 1.0))
	// hit.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, postEmpt, fadeOut, 1.0, 0.0))

	// if len(frames) == 1 {
	// 	if particles {
	// 		hit.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time), fadeOut, 0.9, 1.05))
	// 	} else {
	// 		hit.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time), float64(time+difficulty.ResultFadeIn*0.8), 0.6, 1.1))
	// 		hit.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, fadeIn, float64(time+difficulty.ResultFadeIn*1.2), 1.1, 0.9))
	// 		hit.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time+difficulty.ResultFadeIn*1.2), float64(time+difficulty.ResultFadeIn*1.4), 0.9, 1.0))
	// 	}

	// 	if result == osu.Miss {
	// 		rotation := rand.Float64()*0.3 - 0.15

	// 		hit.AddTransformUnordered(animation.NewSingleTransform(animation.Rotate, easing.Linear, float64(time), fadeIn, 0.0, rotation))
	// 		hit.AddTransformUnordered(animation.NewSingleTransform(animation.Rotate, easing.Linear, fadeIn, fadeOut, rotation, rotation*2))

	// 		hit.AddTransformUnordered(animation.NewSingleTransform(animation.MoveY, easing.Linear, float64(time), fadeOut, position.Y-5, position.Y+40))
	// 	}
	// }

	// hit.SortTransformations()
	// hit.AdjustTimesToTransformations()
	// hit.ResetValuesToTransforms()

	// results.top.Add(hit)

	if (result&osu.BaseHitsM)&(osu.Hit100|osu.Hit50|osu.Miss) > 0 {
		font := font.GetFont("Quicksand Bold")
		text := sprite.NewTextSpriteSize(name, font, font.GetSize()*0.15, float64(time)+1, position, vector.Centre)

		var startPositionBump float64 = 10.0
		if teamIndex == 0 {
			startPositionBump = 10.0
		} else {
			startPositionBump = -10.0
		}
		endPositionBump := startPositionBump * (3 + 2*rand.Float64())

		text.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(time), fadeIn, 0.0, 1.0))
		text.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, postEmpt, fadeOut, 1.0, 0.0))
		text.AddTransformUnordered(animation.NewSingleTransform(animation.MoveY, easing.OutCubic, float64(time), fadeOut, position.Y+startPositionBump, position.Y+endPositionBump))
		text.SortTransformations()
		text.AdjustTimesToTransformations()
		text.ResetValuesToTransforms()

		switch result & osu.BaseHitsM {
		case osu.Hit100:
			text.SetColor(color2.NewRGBA(0.44, 0.98, 0.18, 1))
		case osu.Hit50:
			text.SetColor(color2.NewRGBA(0.2, 0.8, 1, 1))
		case osu.Miss:
			text.SetColor(color2.NewRGBA(0.98, 0.11, 0.011, 1))
		}

		results.top.Add(text)
	}

	// if !settings.Gameplay.ShowHitLighting || result&osu.BaseHitsM < osu.Hit50 {
	// 	return
	// }

	// lighting := sprite.NewSpriteSingle(skin.GetTexture("lighting"), float64(time), position, vector.Centre)
	// lighting.SetColor(skin.GetColor(int(object.GetComboSet()), int(object.GetComboSetHax()), results.color))
	// lighting.SetAdditive(true)
	// lighting.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.OutQuad, float64(time), float64(time+600), 0.8, 1.2))
	// lighting.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(time), float64(time+200), 0, 1))
	// lighting.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(time+400), float64(time+1400), 1, 0))

	// results.bottom.Add(lighting)
}

func (results *HitResults) Update(time float64) {
	results.bottom.Update(time)
	results.top.Update(time)
	results.lastTime = time
}

func (results *HitResults) DrawBottom(batch *batch.QuadBatch, c []color2.Color, alpha float64) {
	results.color = c[0]
	results.alpha = alpha

	batch.ResetTransform()
	batch.SetColor(1, 1, 1, alpha)

	scale := results.diff.CircleRadius / 64
	batch.SetScale(scale, scale)

	results.bottom.Draw(results.lastTime, batch)

	batch.ResetTransform()
}

func (results *HitResults) DrawTop(batch *batch.QuadBatch, _ float64) {
	batch.ResetTransform()
	batch.SetColor(1, 1, 1, results.alpha)

	scale := results.diff.CircleRadius / 64
	batch.SetScale(scale, scale)

	results.top.Draw(results.lastTime, batch)

	batch.ResetTransform()
	batch.SetColor(1, 1, 1, 1)
}
