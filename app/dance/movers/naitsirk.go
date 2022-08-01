package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type NaitsirkMover struct {
	*basicMover

	curve curves.Naitsirk
}

func NewNaitsirkMover() MultiPointMover {
	return &NaitsirkMover{basicMover: &basicMover{}}
}

func getVelocity(v0, x0, x1 vector.Vector2f, dt float32) vector.Vector2f {
	pi := math32.Pi
	pow := math32.Pow
	p1 := x0.Sub(v0).Scl(3*pow(pi, 2) - 32)
	p2 := x1.Sub(x0).Scl(6*pow(pi, 2) - 48)
	div := (dt / 100) * (3*pow(pi, 2) - 16)
	return p1.Add(p2).Scl(1 / div)
}

func (mover *NaitsirkMover) SetObjects(objs []objects.IHitObject) int {
	// config := settings.CursorDance.MoverSettings.Naitsirk[mover.id%len(settings.CursorDance.MoverSettings.Naitsirk)]

	points := make([]vector.Vector2f, 0)
	velocities := make([]vector.Vector2f, 0)
	deltaTimes := make([]float32, 0)

	i := 0

	mover.startTime = objs[0].GetEndTime()
	prevVelocity := objs[0].GetStackedEndPositionMod(mover.diff.Mods)
	for ; i < (len(objs) - 1); i++ {
		o := objs[i]
		oNext := objs[i+1]
		mover.endTime = oNext.GetStartTime()

		if i == 0 {
			oVelocity := prevVelocity.Copy()
			velocities = append(velocities, oVelocity)
			deltaTimes = append(deltaTimes, float32(oNext.GetStartTime()-o.GetEndTime()))
			points = append(points, o.GetStackedEndPositionMod(mover.diff.Mods))

			continue
		}

		oPrev := objs[i-1]
		velChange := getVelocity(prevVelocity,
			oPrev.GetStackedEndPositionMod(mover.diff.Mods),
			oNext.GetStackedStartPositionMod(mover.diff.Mods),
			float32(oNext.GetStartTime()-oPrev.GetEndTime()),
		)
		oVelocity := o.GetStackedStartPositionMod(mover.diff.Mods).Add(velChange)
		prevVelocity = oVelocity.Copy()

		points = append(points, o.GetStackedStartPositionMod(mover.diff.Mods))

		if _, ok := o.(objects.ILongObject); ok {
			mover.endTime = o.GetStartTime()
			velocities = append(velocities, o.GetStackedStartPositionMod(mover.diff.Mods))
			break
		}

		deltaTimes = append(deltaTimes, float32(oNext.GetStartTime()-o.GetEndTime()))
		velocities = append(velocities, oVelocity)
	}

	mover.curve = curves.NewNaitsirk(points, velocities, deltaTimes)

	return i + 1
}

func (mover *NaitsirkMover) Update(time float64) vector.Vector2f {
	t := time - mover.startTime
	return mover.curve.PointAt(float32(t))
}
