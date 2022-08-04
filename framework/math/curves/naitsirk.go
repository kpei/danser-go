package curves

import (
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type Naitsirk struct {
	points     []vector.Vector2f
	velocities []vector.Vector2f
	deltaTimes []float32
}

func NewNaitsirk(points, velocities []vector.Vector2f, deltatimes []float32) Naitsirk {
	dt := make([]float32, 0)
	for _, time := range deltatimes {
		dt = append(dt, time/100)
	}
	return Naitsirk{points, velocities, dt}
}

func S(x0, x1, v0, v1 vector.Vector2f, dt, t float32) vector.Vector2f {
	p1 := v1.Sub(v0).Scl(math32.Pow(t, 2) / (2 * dt))
	p2 := (x1.Sub(x0).Scl(0.5).Sub(v0.Add(v1).Scl(0.25 * dt))).Scl(math32.Cos(math32.Pi * t / dt))
	p3 := v0.Scl(t)
	p4 := v0.Add(v1).Scl(0.25 * dt)
	p5 := x0.Add(x1).Scl(0.5)
	return p1.Sub(p2).Add(p3).Sub(p4).Add(p5)
}

func (c Naitsirk) PointAt(t float32) vector.Vector2f {
	tS := t / 100
	timeElapsed := float32(0)
	for i, deltaTime := range c.deltaTimes {
		if (len(c.points) < 2) || (i >= len(c.points)-1) {
			break
		}

		if (tS >= timeElapsed) && (tS <= (timeElapsed + deltaTime)) {
			x0, x1 := c.points[i], c.points[i+1]
			v0, v1 := c.velocities[i].Sub(c.points[i]), c.velocities[i+1].Sub(c.points[i+1])
			pos := S(x0, x1, v0, v1, deltaTime, tS-timeElapsed)
			return pos
		}
		timeElapsed += deltaTime
	}

	return vector.NewVec2f(100, 100)
}
