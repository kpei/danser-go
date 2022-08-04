package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/vector"
)

type NaitsirkMover struct {
	*basicMover

	curve curves.Naitsirk
}

func NewNaitsirkMover() MultiPointMover {
	return &NaitsirkMover{basicMover: &basicMover{}}
}

func GetVelocity(x0, x1 vector.Vector2f, dt float32) vector.Vector2f {
	div := dt / 100
	return x1.Sub(x0).Scl(1 / div)
}

func (mover *NaitsirkMover) SetObjects(objs []objects.IHitObject) int {
	// config := settings.CursorDance.MoverSettings.Naitsirk[mover.id%len(settings.CursorDance.MoverSettings.Naitsirk)]

	points := make([]vector.Vector2f, 0)
	velocities := make([]vector.Vector2f, 0)
	deltaTimes := make([]float32, 0)

	i := 0

	mover.startTime = objs[0].GetEndTime()
	for ; i < (len(objs) - 1); i++ {
		o := objs[i]
		oNext := objs[i+1]
		mover.endTime = oNext.GetStartTime()

		if i == 0 {
			velocities = append(velocities, objs[0].GetStackedEndPositionMod(mover.diff.Mods))
			deltaTimes = append(deltaTimes, float32(oNext.GetStartTime()-o.GetEndTime()))
			points = append(points, o.GetStackedEndPositionMod(mover.diff.Mods))

			continue
		}

		oPrev := objs[i-1]
		velChange := GetVelocity(
			oPrev.GetStackedEndPositionMod(mover.diff.Mods),
			oNext.GetStackedStartPositionMod(mover.diff.Mods),
			float32(oNext.GetStartTime()-oPrev.GetEndTime()),
		)
		oVelocity := o.GetStackedStartPositionMod(mover.diff.Mods).Add(velChange)

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
