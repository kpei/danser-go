package skills

import (
	"github.com/khezen/rootfinding"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/preprocessing"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"gonum.org/v1/gonum/integrate/quad"
)

type CoordinationSkill struct {
	*Skill
}

func NewCoordinationSkill(d *difficulty.Difficulty, experimental bool) *CoordinationSkill {
	skill := &CoordinationSkill{Skill: NewSkill(d, experimental)}

	skill.SkillMultiplier = 23.25
	skill.StrainDecayBase = 0.1
	skill.HistoryLength = 2
	skill.StrainValueOf = skill.strainValue

	return skill
}

func (skill *CoordinationSkill) strainValue(current *preprocessing.DifficultyObject) float64 {
	if _, ok := current.BaseObject.(*objects.Spinner); ok || len(skill.Previous) <= 1 {
		return 0
	}
	if _, ok := skill.GetPrevious(0).BaseObject.(*objects.Spinner); ok {
		return 0
	}

	coordinationDifficulty := skill.coordinationDifficultyOf(current)
	aimDifficulty := skill.aimDifficultyOf(current)

	return coordinationDifficulty + aimDifficulty
}

func (skill *CoordinationSkill) aimDifficultyOf(current *preprocessing.DifficultyObject) float64 {
	osuToObj := current
	osuFromObj := skill.GetPrevious(0)
	osuNextObj := skill.GetNext(0)
	osuPrevObj := skill.GetPrevious(1)
	if osuFromObj == nil || osuNextObj == nil {
		return 0
	}

	if osuPrevObj == nil {
		osuPrevObj = osuFromObj
	}

	integrand := func(t float64) float64 {
		velAtT := skill.velocityAt(osuPrevObj, osuFromObj, osuToObj, osuNextObj, float32(t))
		distance := float64(math32.Sqrt(velAtT.Dot(velAtT)))
		return distance
	}
	aimDifficulty := quad.Fixed(integrand, 0, osuToObj.StrainTime, 50, nil, 0) / osuToObj.StrainTime

	return aimDifficulty
}

func (skill *CoordinationSkill) coordinationDifficultyOf(current *preprocessing.DifficultyObject) float64 {
	osuCurrObj := current
	osuLastObj := skill.GetPrevious(0)
	osuNextObj := skill.GetNext(0)
	osuNextNextObj := skill.GetNext(1)

	timeInNote := 0.0

	// The circle located at (X, Y) with radius r can be described by the equation (x - X)^2 + (y - Y)^2 = r^2.
	// We can find when the position function intersects the circle by substituting xComponent(t) into x and yComponent(t) into y,
	// subtracting r^2 from both sides of the equation, and then solving for t.
	// Because the positions are normalized with respect to the radius, r^2 = 1.
	root := func(x0, x1, x2, target *preprocessing.DifficultyObject, start bool, t float64) float64 {
		cursor := skill.positionAt(x0, x1, x2, float32(t))
		position := target.NormalizedEndPosition
		if start {
			position = target.NormalizedStartPosition
		}
		distance := cursor.Sub(position)
		return float64(distance.Dot(distance) - 1)
	}

	// Determine the amount of time the cursor is within the current circle as it moves from the previous circle.
	if osuLastObj != nil && osuNextObj != nil {
		// This distance must be computed because sliders aren't being taken into account.
		distance := osuCurrObj.NormalizedStartPosition.Sub(osuLastObj.NormalizedEndPosition)
		realSquaredDistance := distance.Dot(distance)

		// If the current and previous objects are overlapped by 50% or more, just add the DeltaTime of the current object.
		if realSquaredDistance <= 1 {
			timeInNote += osuCurrObj.StrainTime
		} else {
			timeEnterNote, _ := rootfinding.Brent(func(t float64) float64 { return root(osuLastObj, osuCurrObj, osuNextObj, osuCurrObj, true, t) }, 0, osuCurrObj.StrainTime, 3)
			timeInNote += osuCurrObj.StrainTime - timeEnterNote
		}
	} else {
		timeInNote += skill.diff.Hit50U
	}

	// Determine the amount of time the cursor is within the current circle as it moves toward the next circle.
	if osuNextNextObj != nil && osuNextObj != nil {
		distance := osuNextObj.NormalizedStartPosition.Sub(osuCurrObj.NormalizedEndPosition)
		realSquaredDistance := distance.Dot(distance)

		if realSquaredDistance <= 1 {
			timeInNote += osuNextObj.StrainTime
		} else {
			timeExitNote, _ := rootfinding.Brent(func(t float64) float64 { return root(osuCurrObj, osuNextObj, osuNextNextObj, osuCurrObj, false, t) }, 0, osuNextObj.StrainTime, 3)
			timeInNote += timeExitNote
		}
	} else {
		timeInNote += skill.diff.Hit50U
	}

	return 2 / timeInNote
}

func (skill *CoordinationSkill) positionAt(x0, x1, x2 *preprocessing.DifficultyObject, t float32) vector.Vector2f {
	deltaTime := float32(x1.StrainTime)
	velocityDeltaTime := float32(x2.StartTime - x0.EndTime)
	v0 := x0.NormalizedEndPosition.Sub(x0.NormalizedEndPosition)
	v1 := x1.NormalizedStartPosition.Add(x2.NormalizedStartPosition.Sub(x0.NormalizedEndPosition).Scl(1 / velocityDeltaTime)).Sub(x1.NormalizedStartPosition)

	return curves.S(x0.NormalizedEndPosition, x1.NormalizedStartPosition, v0, v1, deltaTime, t)
}

func (skill *CoordinationSkill) velocityAt(x0, x1, x2, x3 *preprocessing.DifficultyObject, t float32) vector.Vector2f {
	pi := math32.Pi

	deltaTime := float32(x2.StrainTime)
	prevVelocityDeltaTime := float32(x2.StartTime - x0.EndTime)
	velocityDeltaTime := float32(x3.StartTime - x1.EndTime)
	v0 := x1.NormalizedEndPosition.Add(x2.NormalizedStartPosition.Sub(x0.NormalizedEndPosition).Scl(1 / prevVelocityDeltaTime)).Sub(x1.NormalizedEndPosition)
	v1 := x2.NormalizedStartPosition.Add(x3.NormalizedStartPosition.Sub(x1.NormalizedEndPosition).Scl(1 / velocityDeltaTime)).Sub(x2.NormalizedStartPosition)

	p1 := v1.Sub(v0).Scl(t / deltaTime)
	p2 := v0.Add(v1).Scl(deltaTime).Add(x0.NormalizedEndPosition.Sub(x1.NormalizedEndPosition).Scl(2))
	p3 := p2.Scl(pi * math32.Sin(pi*t/deltaTime) / (4 * deltaTime))

	return v0.Add(p1).Sub(p3)
}
