package play

import (
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/shape"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type ScoreBar struct {
	scoreTeam1 int64
	scoreTeam2 int64
	pctTeam1   float32

	scoreBarGlider *animation.Glider
	shapeRenderer  *shape.Renderer
	font           *font.Font
}

func NewScoreBar(font *font.Font) *ScoreBar {
	return &ScoreBar{
		scoreTeam1:     1,
		scoreTeam2:     1,
		pctTeam1:       0.5,
		scoreBarGlider: animation.NewGlider(0.5),
		shapeRenderer:  shape.NewRenderer(),
		font:           font,
	}
}

func (scoreBar *ScoreBar) NextScore(time float64) {
	scoreBar.scoreBarGlider.AddEvent(time, time+750, float64(scoreBar.pctTeam1))
}

func (scoreBar *ScoreBar) Update(scoreTeam1 int64, scoreTeam2 int64, time float64) {
	scoreBar.scoreTeam1 = scoreTeam1
	scoreBar.scoreTeam2 = scoreTeam2
	scoreBar.pctTeam1 = math32.Max(float32(scoreTeam1), 1) / math32.Max(float32(scoreTeam1+scoreTeam2), 2)
	scoreBar.scoreBarGlider.Update(time)
}

func (scoreBar *ScoreBar) Draw(batch *batch.QuadBatch, alpha float64) {
	batch.Flush()

	const thickness float32 = 40
	const yPos float32 = thickness / 2
	var scaledWidth float32 = float32(settings.Graphics.GetWidth())
	var xPos float32 = scaledWidth * float32(scoreBar.scoreBarGlider.GetValue())

	scoreBar.shapeRenderer.Begin()
	scoreBar.shapeRenderer.SetCamera(batch.Projection)

	scoreBar.shapeRenderer.SetColor(0.18, 0.525, 0.87, alpha*0.75)
	scoreBar.shapeRenderer.DrawLineV(vector.NewVec2f(0, yPos), vector.NewVec2f(xPos, yPos), thickness)
	scoreBar.shapeRenderer.SetColor(0.93, 0.32, 0.33, alpha*0.75)
	scoreBar.shapeRenderer.DrawLineV(vector.NewVec2f(xPos, yPos), vector.NewVec2f(scaledWidth, yPos), thickness)

	scoreBar.shapeRenderer.End()

	scoreTeam1Str := utils.Humanize(scoreBar.scoreTeam1)
	scoreTeam2Str := utils.Humanize(scoreBar.scoreTeam2)
	scoreTeam1FontWidth := scoreBar.font.GetWidth(24, scoreTeam1Str)
	batch.SetColor(1, 1, 1, alpha*0.75)
	scoreBar.font.DrawOrigin(batch, float64(xPos)-scoreTeam1FontWidth-25, float64(yPos), vector.CentreLeft, 24, false, scoreTeam1Str)
	scoreBar.font.DrawOrigin(batch, float64(xPos)+25, float64(yPos), vector.CentreLeft, 24, false, scoreTeam2Str)
}
