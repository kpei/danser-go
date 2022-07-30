package overlays

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strconv"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/dance"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/rulesets/osu/performance"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/app/states/components/common"
	"github.com/wieku/danser-go/app/states/components/overlays/play"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
)

type stats struct {
	pp       float64
	score    int64
	accuracy float64
}

type knockoutPlayer struct {
	fade   *animation.Glider
	slide  *animation.Glider
	height *animation.Glider
	index  *animation.Glider

	scoreDisp *animation.TargetGlider
	ppDisp    *animation.TargetGlider
	accDisp   *animation.TargetGlider

	lastCombo int64
	sCombo    int64
	maxCombo  int64
	hasBroken bool
	breakTime int64
	pp        float64
	score     int64

	perObjectStats []stats

	displayHp float64

	lastHit  osu.HitResult
	fadeHit  *animation.Glider
	scaleHit *animation.Glider

	avatarCircle *play.AvatarCircle

	name         string
	oldIndex     int
	currentIndex int
}

type bubble struct {
	deathFade  *animation.Glider
	deathSlide *animation.Glider
	deathX     float64
	endTime    float64
	name       string
	combo      int64
	lastHit    osu.HitResult
	lastCombo  osu.ComboResult
	deathScale *animation.Glider
}

func newBubble(position vector.Vector2d, time float64, name string, combo int64, lastHit osu.HitResult, lastCombo osu.ComboResult) *bubble {
	deathShiftX := (rand.Float64() - 0.5) * 10
	deathShiftY := (rand.Float64() - 0.5) * 10
	baseY := position.Y + deathShiftY

	bub := new(bubble)
	bub.name = name
	bub.deathX = position.X + deathShiftX
	bub.deathSlide = animation.NewGlider(0.0)
	bub.deathFade = animation.NewGlider(0.0)
	bub.deathScale = animation.NewGlider(1)
	bub.deathSlide.SetEasing(easing.OutQuad)

	if settings.Knockout.Mode == settings.OneVsOne {
		bub.deathSlide.AddEventS(time, time+2000, baseY, baseY)
		bub.deathFade.AddEventS(time, time+difficulty.ResultFadeIn, 0, 1)
		bub.deathFade.AddEventS(time+difficulty.PostEmpt, time+difficulty.PostEmpt+difficulty.ResultFadeOut, 1, 0)
		bub.deathScale.AddEventSEase(time, time+difficulty.ResultFadeIn*1.2, 0.4, 1, easing.OutElastic)
	} else {
		bub.deathSlide.AddEventS(time, time+2000, baseY, baseY+50)
		bub.deathFade.AddEventS(time, time+200, 0, 1)
		bub.deathFade.AddEventS(time+800, time+1200, 1, 0)
	}

	bub.endTime = time + 2000
	bub.combo = combo
	bub.lastHit = lastHit
	bub.lastCombo = lastCombo

	return bub
}

type missTimeline struct {
	playerName string
	result     osu.HitResult
	progress   float64
}

type KnockoutOverlay struct {
	controller   *dance.ReplayController
	font         *font.Font
	players      map[string]*knockoutPlayer
	playersArray []*knockoutPlayer
	deathBubbles []*bubble
	scorebar     *play.ScoreBar
	names        map[*graphics.Cursor]string
	generator    *rand.Rand

	results      *play.HitResults
	strainGraph  *play.StrainGraph
	missTimeline []*missTimeline

	audioTime  float64
	normalTime float64

	boundaries *common.Boundaries

	Button        *texture.TextureRegion
	ButtonClicked *texture.TextureRegion

	ScaledHeight float64
	ScaledWidth  float64

	music bass.ITrack

	breakMode bool
	fade      *animation.Glider
}

func NewKnockoutOverlay(replayController *dance.ReplayController) *KnockoutOverlay {
	overlay := new(KnockoutOverlay)
	overlay.controller = replayController

	if font.GetFont("Quicksand Bold") == nil {
		file, _ := assets.Open("assets/fonts/Quicksand-Bold.ttf")
		font.LoadFont(file)
		file.Close()
	}

	overlay.font = font.GetFont("Quicksand Bold")

	overlay.missTimeline = make([]*missTimeline, 0)
	overlay.players = make(map[string]*knockoutPlayer)
	overlay.playersArray = make([]*knockoutPlayer, 0)
	overlay.deathBubbles = make([]*bubble, 0)
	overlay.names = make(map[*graphics.Cursor]string)
	overlay.generator = rand.New(rand.NewSource(replayController.GetBeatMap().TimeAdded))
	overlay.scorebar = play.NewScoreBar(overlay.font)

	overlay.ScaledHeight = 1080.0
	overlay.ScaledWidth = overlay.ScaledHeight * settings.Graphics.GetAspectRatio()

	overlay.strainGraph = play.NewStrainGraph(replayController.GetRuleset())

	overlay.fade = animation.NewGlider(1)

	overlay.results = play.NewHitResults(replayController.GetRuleset().GetBeatMap().Diff)

	for i, r := range replayController.GetReplays() {
		teamColor := color.RGBA{46, 134, 222, 255}
		if i == 1 {
			teamColor = color.RGBA{238, 82, 83, 255}
		}

		overlay.names[replayController.GetCursors()[i]] = r.Name
		overlay.players[r.Name] = &knockoutPlayer{animation.NewGlider(1), animation.NewGlider(0), animation.NewGlider(overlay.ScaledHeight * 0.9 * 1.04 / (51)), animation.NewGlider(float64(i)), animation.NewTargetGlider(0, 0), animation.NewTargetGlider(0, 2), animation.NewTargetGlider(100, 2), 0, 0, r.MaxCombo, false, 0, 0.0, 0, make([]stats, len(replayController.GetBeatMap().HitObjects)), 0.0, osu.Hit300, animation.NewGlider(0), animation.NewGlider(0), play.NewAvatarCircle(r.Name, teamColor), r.Name, i, i}
		overlay.players[r.Name].index.SetEasing(easing.InOutQuad)
		overlay.playersArray = append(overlay.playersArray, overlay.players[r.Name])
	}

	discord.UpdateKnockout(len(overlay.playersArray), len(overlay.playersArray))

	for i, g := range overlay.playersArray {
		if i != g.currentIndex {
			g.index.Reset()
			g.index.SetValue(float64(i))
			g.currentIndex = i
		}
	}

	replayController.GetRuleset().SetListener(overlay.hitReceived)

	replayController.GetRuleset().SetEndListener(func(time int64, number int64) {
	})

	overlay.boundaries = common.NewBoundaries()

	overlay.Button = skin.GetTexture("knockout-button")
	overlay.ButtonClicked = skin.GetTexture("knockout-button-active")

	return overlay
}

func (overlay *KnockoutOverlay) hitReceived(cursor *graphics.Cursor, time int64, number int64, position vector.Vector2d, result osu.HitResult, comboResult osu.ComboResult, ppResults performance.PPv2Results, score int64) {
	if result == osu.PositionalMiss {
		return
	}

	player := overlay.players[overlay.names[cursor]]

	if overlay.controller.GetRuleset().GetBeatMap().Diff.Mods.Active(difficulty.HardRock) != overlay.controller.GetReplays()[player.oldIndex].ModsV.Active(difficulty.HardRock) {
		position.Y = 384 - position.Y
	}

	player.score = score
	player.pp = ppResults.Total

	player.scoreDisp.SetValue(float64(score), false)
	player.ppDisp.SetValue(player.pp, false)

	sc := overlay.controller.GetRuleset().GetScore(cursor)

	player.perObjectStats[number].score = score
	player.perObjectStats[number].pp = ppResults.Total
	player.perObjectStats[number].accuracy = sc.Accuracy

	player.accDisp.SetValue(sc.Accuracy, false)

	if comboResult == osu.Increase {
		player.sCombo++
	}

	resultClean := result & osu.BaseHitsM

	acceptableHits := resultClean&(osu.Hit100|osu.Hit50|osu.Miss) > 0
	if acceptableHits {
		player.fadeHit.Reset()
		player.fadeHit.AddEventS(overlay.normalTime, overlay.normalTime+300, 0.5, 1)
		player.fadeHit.AddEventS(overlay.normalTime+600, overlay.normalTime+900, 1, 0)
		player.scaleHit.AddEventS(overlay.normalTime, overlay.normalTime+300, 0.5, 1)
		player.lastHit = result & (osu.HitValues | osu.Miss) //resultClean
		// if settings.Knockout.Mode == settings.OneVsOne {
		// 	overlay.deathBubbles = append(overlay.deathBubbles, newBubble(position, overlay.normalTime, overlay.names[cursor], player.sCombo, resultClean, comboResult))
		// }
	}

	object := overlay.controller.GetRuleset().GetBeatMap().HitObjects[number]
	if resultClean > 0 {
		overlay.results.AddResult(time, result, player.name, player.currentIndex, position, object)
	}

	comboBreak := comboResult == osu.Reset
	if (settings.Knockout.Mode == settings.SSOrQuit && (acceptableHits || comboBreak)) || (comboBreak && number != 0) {

		if !player.hasBroken {
			if settings.Knockout.Mode == settings.XReplays {
				if player.sCombo >= int64(settings.Knockout.BubbleMinimumCombo) {
					// overlay.deathBubbles = append(overlay.deathBubbles, newBubble(position, overlay.normalTime, overlay.names[cursor], player.sCombo, resultClean, comboResult))
					log.Println(overlay.names[cursor], "has broken! Combo:", player.sCombo)
				}
			} else if settings.Knockout.Mode == settings.SSOrQuit ||
				(settings.Knockout.Mode == settings.ComboBreak && time > int64(settings.Knockout.GraceEndTime*1000)) ||
				(settings.Knockout.Mode == settings.MaxCombo && math.Abs(float64(player.sCombo-player.maxCombo)) < 5) {
				//Fade out player name
				player.hasBroken = true
				player.breakTime = time

				player.fade.AddEvent(overlay.normalTime, overlay.normalTime+3000, 0)

				player.height.SetEasing(easing.OutQuad)
				player.height.AddEvent(overlay.normalTime+2500, overlay.normalTime+3000, 0)

				overlay.deathBubbles = append(overlay.deathBubbles, newBubble(position, overlay.normalTime, overlay.names[cursor], player.sCombo, resultClean, comboResult))

				log.Println(overlay.names[cursor], "has broken! Max combo:", player.sCombo)
			}
		}
	}

	if comboBreak {
		player.sCombo = 0
	}

	if (resultClean) > 0 {
		// Some shit here for the miss timeline
		missTimeline := &missTimeline{
			playerName: player.name,
			result:     result,
			progress:   overlay.strainGraph.Progress(),
		}
		overlay.missTimeline = append(overlay.missTimeline, missTimeline)
	}

	overlay.scorebar.NextScore(overlay.normalTime)
}

func (overlay *KnockoutOverlay) Update(time float64) {
	if overlay.audioTime == 0 {
		overlay.audioTime = time
		overlay.normalTime = time
	}

	delta := time - overlay.audioTime

	if overlay.music != nil && overlay.music.GetState() == bass.MusicPlaying {
		delta /= overlay.music.GetTempo()
	}

	overlay.normalTime += delta

	overlay.audioTime = time

	team1Score := overlay.playersArray[0].score
	team2Score := overlay.playersArray[1].score
	overlay.scorebar.Update(team1Score, team2Score, overlay.normalTime)

	overlay.results.Update(time)
	overlay.updateBreaks(overlay.normalTime)
	overlay.fade.Update(overlay.normalTime)
	overlay.strainGraph.Update(overlay.audioTime)

	for _, r := range overlay.controller.GetReplays() {
		player := overlay.players[r.Name]
		player.height.Update(overlay.normalTime)
		player.fade.Update(overlay.normalTime)
		player.fadeHit.Update(overlay.normalTime)
		player.scaleHit.Update(overlay.normalTime)
		player.index.Update(overlay.normalTime)
		player.scoreDisp.Update(overlay.normalTime)
		player.ppDisp.Update(overlay.normalTime)
		player.accDisp.Update(overlay.normalTime)
		player.lastCombo = r.Combo

		currentHp := overlay.controller.GetRuleset().GetHP(overlay.controller.GetCursors()[player.oldIndex])

		if player.displayHp < currentHp {
			player.displayHp = math.Min(1.0, player.displayHp+math.Abs(currentHp-player.displayHp)/4*delta/16.667)
		} else if player.displayHp > currentHp {
			player.displayHp = math.Max(0.0, player.displayHp-math.Abs(player.displayHp-currentHp)/6*delta/16.667)
		}
	}
}

func (overlay *KnockoutOverlay) SetMusic(music bass.ITrack) {
	overlay.music = music
}

func (overlay *KnockoutOverlay) DrawBackground(batch *batch.QuadBatch, _ []color2.Color, alpha float64) {
	alpha *= overlay.fade.GetValue()
	overlay.boundaries.Draw(batch.Projection, float32(overlay.controller.GetBeatMap().Diff.CircleRadius), float32(alpha))
}

func (overlay *KnockoutOverlay) DrawBeforeObjects(batch *batch.QuadBatch, c []color2.Color, alpha float64) {
	overlay.results.DrawBottom(batch, c, alpha)
}

func (overlay *KnockoutOverlay) DrawNormal(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
	alpha *= overlay.fade.GetValue()

	overlay.results.DrawTop(batch, 1.0)

	batch.ResetTransform()

	alive := 0
	for _, r := range overlay.controller.GetReplays() {
		player := overlay.players[r.Name]
		if !player.hasBroken {
			alive++
		}
	}

	minSize := settings.Knockout.MinCursorSize
	maxSize := settings.Knockout.MaxCursorSize
	settings.Cursor.CursorSize = minSize + (maxSize-minSize)*math.Pow(1-math.Sin(float64(alive)/math.Max(51, float64(settings.PLAYERS))*math.Pi/2), 3)

	batch.SetScale(1, 1)
}

func (overlay *KnockoutOverlay) DrawHUD(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
	alpha *= overlay.fade.GetValue()

	batch.ResetTransform()

	controller := overlay.controller
	replays := controller.GetReplays()

	scl := overlay.ScaledHeight * 0.9 / 51
	//margin := scl*0.02

	highestCombo := int64(0)
	highestPP := 0.0
	highestACC := 0.0
	highestScore := int64(0)
	cumulativeHeight := 0.0
	maxPlayerWidth := 0.0

	overlay.scorebar.Draw(batch, alpha)

	for _, r := range replays {
		cumulativeHeight += overlay.players[r.Name].height.GetValue()

		highestCombo = mutils.Max(highestCombo, overlay.players[r.Name].sCombo)
		highestPP = math.Max(highestPP, overlay.players[r.Name].pp)
		highestACC = math.Max(highestACC, r.Accuracy)
		highestScore = mutils.Max(highestScore, overlay.players[r.Name].score)

		pWidth := overlay.font.GetWidth(scl, r.Name)

		if r.Mods != "" {
			pWidth += overlay.font.GetWidth(scl*0.8, "+"+r.Mods)
		}

		maxPlayerWidth = math.Max(maxPlayerWidth, pWidth)
	}

	cL := strconv.FormatInt(highestCombo, 10)
	// cP := strconv.FormatInt(int64(highestPP), 10)
	cA := strconv.FormatInt(int64(highestACC), 10)
	// cS := overlay.font.GetWidthMonospaced(scl, utils.Humanize(highestScore))

	accuracy1 := cA + ".00% "
	nWidth := overlay.font.GetWidthMonospaced(scl, accuracy1)
	combo1 := cL + "x"
	nWidthCL := overlay.font.GetWidthMonospaced(scl, combo1)

	// maxLength := 3.2*scl + nWidth + maxPlayerWidth

	// xSlideLeft := (overlay.fade.GetValue() - 1.0) * maxLength
	// xSlideRight := (1.0 - overlay.fade.GetValue()) * (cS + overlay.font.GetWidthMonospaced(scl, fmt.Sprintf("%dx ", highestCombo)) + 0.5*scl)

	rowPosY := math.Max((overlay.ScaledHeight-cumulativeHeight)/2, scl)
	avatarCircleX := 100.0
	avatarCircleY := 120.0
	// Draw textures like keys, grade, hit values
	for i, rep := range overlay.playersArray {
		r := replays[rep.oldIndex]
		player := overlay.players[r.Name]

		// rowBaseY := rowPosY + rep.index.GetValue()*(overlay.ScaledHeight*0.9*1.04/(51)) + player.height.GetValue()/2 /*+margin*10*/
		rowPosY -= overlay.ScaledHeight*0.9*1.04/(51) - player.height.GetValue()

		//batch.SetColor(0.1, 0.8, 0.4, alpha*player.fade.GetValue()*0.4)
		//add := 0.3 + float64(int(math.Round(rep.index.GetValue()))%2)*0.2
		//batch.SetColor(add, add, add, alpha*player.fade.GetValue()*0.7)
		//batch.SetAdditive(true)
		//batch.SetSubScale(player.displayHp*30.5*scl*0.9/2, scl*0.9/2)
		//batch.SetTranslation(vector.NewVec2d(player.displayHp*30.5/2*scl*0.9/2 /*rowPosY*/, rowBaseY))
		//batch.DrawUnit(graphics.Pixel.GetRegion())
		//batch.SetSubScale(16.5*scl*0.9/2, scl*0.9/2)
		//batch.SetTranslation(vector.NewVec2d(settings.Graphics.GetWidthF()-16.5/2*scl*0.9/2 /*rowPosY*/, rowBaseY))
		//batch.DrawUnit(graphics.Pixel.GetRegion())
		//batch.SetAdditive(false)

		batch.SetColor(float64(colors[rep.oldIndex].R), float64(colors[rep.oldIndex].G), float64(colors[rep.oldIndex].B), alpha*player.fade.GetValue())

		for j := 0; j < 2; j++ {
			batch.SetSubScale(scl*0.8/2, scl*0.8/2)
			if i == 1 {
				batch.SetTranslation(vector.NewVec2d((overlay.ScaledWidth-avatarCircleX)-player.avatarCircle.GetWidth()-5+float64(j)*scl, 150))
			} else {
				batch.SetTranslation(vector.NewVec2d(avatarCircleX+player.avatarCircle.GetWidth()+5-float64(j)*scl, 150))
			}

			if controller.GetClick(rep.oldIndex, j) || controller.GetClick(rep.oldIndex, j+2) {
				batch.DrawUnit(*overlay.ButtonClicked)
			} else {
				batch.DrawUnit(*overlay.Button)
			}
		}

		width := overlay.font.GetWidth(scl, r.Name)

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		if r.Mods != "" {
			width += overlay.font.GetWidth(scl*0.8, "+"+r.Mods)
		}

		if r.Grade != osu.NONE {
			text := skin.GetTexture("ranking-" + r.Grade.TextureName() + "-small")

			ratio := 1.0 / 44.0 // default skin's grade height
			if text.Height < 44 {
				ratio = 1.0 / float64(text.Height) // if skin's grade is smaller, make it bigger
			}

			batch.SetSubScale(scl*0.9*ratio, scl*0.9*ratio)
			if i == 1 {
				batch.SetTranslation(vector.NewVec2d((overlay.ScaledWidth-avatarCircleX)-player.avatarCircle.GetWidth()-5-1.5*scl, 150))
			} else {
				batch.SetTranslation(vector.NewVec2d(avatarCircleX+player.avatarCircle.GetWidth()+5+1.5*scl, 150))
			}

			batch.DrawTexture(*text)
		}

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue()*player.fadeHit.GetValue())
		//batch.SetSubScale(scl*0.9/2*player.scaleHit.GetValue(), scl*0.9/2*player.scaleHit.GetValue())
		//batch.SetTranslation(vector.NewVec2d(3*scl+width+nWidth+scl*0.5, rowBaseY))

		if player.lastHit != 0 {
			tex := ""

			switch player.lastHit & osu.BaseHitsM {
			case osu.Hit300:
				tex = "hit300"
			case osu.Hit100:
				tex = "hit100"
			case osu.Hit50:
				tex = "hit50"
			case osu.Miss:
				tex = "hit0"
			}

			switch player.lastHit & osu.Additions {
			case osu.KatuAddition:
				tex += "k"
			case osu.GekiAddition:
				tex += "g"
			}

			if tex != "" {
				hitTexture := skin.GetTexture(tex)
				batch.SetSubScale(scl*0.9*player.scaleHit.GetValue()*(float64(hitTexture.Width)/float64(hitTexture.Height)), scl*0.9*player.scaleHit.GetValue())
				if i == 1 {
					batch.SetTranslation(vector.NewVec2d((overlay.ScaledWidth-avatarCircleX)-player.avatarCircle.GetWidth()-5-3.6*scl-nWidth-nWidthCL, 150))
				} else {
					batch.SetTranslation(vector.NewVec2d(avatarCircleX+player.avatarCircle.GetWidth()+5+3.6*scl+nWidth+nWidthCL, 150))
				}
				batch.DrawUnit(*hitTexture)
			}
		}
	}

	batch.ResetTransform()

	overlay.strainGraph.Draw(batch, alpha)

	// Draw misstimeline
	strainGraphSize := vector.NewVec2d(settings.Gameplay.StrainGraph.Width, settings.Gameplay.StrainGraph.Height)
	strainGraphXPos := settings.Gameplay.StrainGraph.XPosition
	strainGraphYPos := settings.Gameplay.StrainGraph.YPosition
	for _, miss := range overlay.missTimeline {
		tex := ""
		switch miss.result & osu.BaseHitsM {
		case osu.Hit100:
			tex = "hit100"
		case osu.Hit50:
			tex = "hit50"
		case osu.Miss:
			tex = "hit0"
		}
		if tex != "" {
			hitTexture := skin.GetTexture(tex)
			playerAvatarCircle := overlay.players[miss.playerName].avatarCircle
			x := strainGraphXPos + (strainGraphSize.X * miss.progress)
			y := strainGraphYPos - float64(overlay.strainGraph.StrainY(miss.progress))*strainGraphSize.Y

			batch.ResetTransform()
			batch.SetSubScale(0.2, 0.2)
			playerAvatarCircle.Draw(x, y, alpha, batch)

			batch.SetSubScale(scl*0.9, scl*0.9)
			batch.SetColor(1, 1, 1, alpha)
			batch.SetTranslation(vector.NewVec2d(x, y-20))
			batch.DrawUnit(*hitTexture)
		}
	}

	batch.ResetTransform()

	rowPosY = math.Max((overlay.ScaledHeight-cumulativeHeight)/2, scl)
	// ascScl := overlay.font.GetAscent() * (scl / overlay.font.GetSize()) / 2

	// Draw texts
	for i, rep := range overlay.playersArray {
		r := replays[rep.oldIndex]
		player := overlay.players[r.Name]

		// rowBaseY := rowPosY + rep.index.GetValue()*(overlay.ScaledHeight*0.9*1.04/(51)) + player.height.GetValue()/2 /*+margin*10*/
		rowPosY -= overlay.ScaledHeight*0.9*1.04/(51) - player.height.GetValue()

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		if i == 1 {
			overlay.font.DrawOrigin(batch, (overlay.ScaledWidth-avatarCircleX)-player.avatarCircle.GetWidth()+25, avatarCircleY, vector.BottomRight, scl*2.4, false, r.Name)
			player.avatarCircle.Draw(overlay.ScaledWidth-avatarCircleX, avatarCircleY, alpha, batch)
		} else {
			overlay.font.DrawOrigin(batch, avatarCircleX+player.avatarCircle.GetWidth()-25, avatarCircleY, vector.BottomLeft, scl*2.4, false, r.Name)
			player.avatarCircle.Draw(avatarCircleX, avatarCircleY, alpha, batch)
		}

		accuracy := fmt.Sprintf("%.2f%%", overlay.players[r.Name].accDisp.GetValue())
		//_ = cL

		if i == 1 {
			overlay.font.DrawOrigin(batch, (overlay.ScaledWidth-avatarCircleX)-player.avatarCircle.GetWidth()-5-2.4*scl, 150, vector.CentreRight, scl, true, accuracy)
		} else {
			overlay.font.DrawOrigin(batch, avatarCircleX+player.avatarCircle.GetWidth()+5+2.4*scl, 150, vector.CentreLeft, scl, true, accuracy)
		}

		sWC := fmt.Sprintf("%dx ", overlay.players[r.Name].sCombo)

		if i == 1 {
			overlay.font.DrawOrigin(batch, (overlay.ScaledWidth-avatarCircleX)-player.avatarCircle.GetWidth()-5-2.35*scl-nWidth, 150, vector.CentreRight, scl, true, sWC)
		} else {
			overlay.font.DrawOrigin(batch, avatarCircleX+player.avatarCircle.GetWidth()+5+2.35*scl+nWidth, 150, vector.CentreLeft, scl, true, sWC)
		}

		// batch.SetColor(float64(colors[rep.oldIndex].R), float64(colors[rep.oldIndex].G), float64(colors[rep.oldIndex].B), alpha*player.fade.GetValue())
		// overlay.font.DrawOrigin(batch, 3.2*scl+nWidth+xSlideLeft, rowBaseY, vector.CentreLeft, scl, false, r.Name)
		// width := overlay.font.GetWidth(scl, r.Name)

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		// TODO: MODS
		// if r.Mods != "" {
		// 	overlay.font.DrawOrigin(batch, 3.2*scl+width+nWidth+xSlideLeft, rowBaseY+ascScl, vector.BottomLeft, scl*0.8, false, "+"+r.Mods)
		// }
	}
}

func (overlay *KnockoutOverlay) IsBroken(cursor *graphics.Cursor) bool {
	return overlay.players[overlay.names[cursor]].hasBroken
}

func (overlay *KnockoutOverlay) updateBreaks(time float64) {
	inBreak := false

	for _, b := range overlay.controller.GetRuleset().GetBeatMap().Pauses {
		if overlay.audioTime < b.GetStartTime() {
			break
		}

		if b.GetEndTime()-b.GetStartTime() >= 1000 && overlay.audioTime >= b.GetStartTime() && overlay.audioTime <= b.GetEndTime() {
			inBreak = true

			break
		}
	}

	if !overlay.breakMode && inBreak {
		if settings.Knockout.HideOverlayOnBreaks {
			overlay.fade.AddEventEase(time, time+500, 0, easing.OutQuad)
		}
	} else if overlay.breakMode && !inBreak {
		overlay.fade.AddEventEase(time, time+500, 1, easing.OutQuad)
	}

	overlay.breakMode = inBreak
}

func (overlay *KnockoutOverlay) DisableAudioSubmission(_ bool) {}

func (overlay *KnockoutOverlay) ShouldDrawHUDBeforeCursor() bool {
	return false
}
