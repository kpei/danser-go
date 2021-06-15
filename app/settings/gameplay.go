package settings

var Gameplay = initGameplay()

func initGameplay() *gameplay {
	return &gameplay{
		HitErrorMeter: &hitError{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			ShowUnstableRate:     true,
			UnstableRateDecimals: 0,
			UnstableRateScale:    1.0,
		},
		Score: &score{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			ProgressBar:     "Pie",
			ShowGradeAlways: false,
		},
		HpBar: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		ComboCounter: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		PPCounter: &ppCounter{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			XPosition:     5,
			YPosition:     150,
			Decimals:      0,
			Align:         "CentreLeft",
			ShowInResults: true,
		},
		HitCounter: &hitCounter{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			Color: []*hsv{
				{
					Hue:        0,
					Saturation: 0,
					Value:      1,
				},
			},
			XPosition:  5,
			YPosition:  190,
			Spacing:    48,
			FontScale:  1,
			Align:      "Left",
			ValueAlign: "Left",
			Show300:    false,
		},
		KeyOverlay: &hudElement{
			Show:    true,
			Scale:   1.0,
			Opacity: 1.0,
		},
		ScoreBoard: &scoreBoard{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			HideOthers:  false,
			ShowAvatars: false,
			YOffset:     0,
		},
		Mods: &mods{
			hudElement: &hudElement{
				Show:    true,
				Scale:   1.0,
				Opacity: 1.0,
			},
			HideInReplays: false,
			FoldInReplays: false,
		},
		Boundaries: &boundaries{
			Enabled:         true,
			BorderThickness: 1,
			BorderFill:      1,
			BorderColor: &hsv{
				Hue:        0,
				Saturation: 0,
				Value:      1,
			},
			BorderOpacity: 1,
			BackgroundColor: &hsv{
				Hue:        0,
				Saturation: 1,
				Value:      0,
			},
			BackgroundOpacity: 0.5,
		},
		ShowResultsScreen: true,
		ResultsScreenTime: 5,
		ShowWarningArrows: true,
		FlashlightDim:     1,
		PlayUsername:      "Guest",
	}
}

type gameplay struct {
	HitErrorMeter     *hitError
	Score             *score
	HpBar             *hudElement
	ComboCounter      *hudElement
	PPCounter         *ppCounter
	HitCounter        *hitCounter
	KeyOverlay        *hudElement
	ScoreBoard        *scoreBoard
	Mods              *mods
	Boundaries        *boundaries
	ShowResultsScreen bool
	ResultsScreenTime float64
	ShowWarningArrows bool
	FlashlightDim     float64
	PlayUsername      string
}

type boundaries struct {
	Enabled bool

	BorderThickness float64
	BorderFill      float64

	BorderColor   *hsv
	BorderOpacity float64

	BackgroundColor   *hsv
	BackgroundOpacity float64
}

type hudElement struct {
	Show    bool
	Scale   float64
	Opacity float64
}

type hitError struct {
	*hudElement
	ShowUnstableRate     bool
	UnstableRateDecimals int
	UnstableRateScale    float64
}

type score struct {
	*hudElement
	ProgressBar     string
	ShowGradeAlways bool
}

type ppCounter struct {
	*hudElement
	XPosition     float64
	YPosition     float64
	Decimals      int
	Align         string
	ShowInResults bool
}

type hitCounter struct {
	*hudElement
	Color      []*hsv
	XPosition  float64
	YPosition  float64
	Spacing    float64
	FontScale  float64
	Align      string
	ValueAlign string
	Show300    bool
}

type scoreBoard struct {
	*hudElement
	HideOthers  bool
	ShowAvatars bool
	YOffset     float64
}

type mods struct {
	*hudElement
	HideInReplays bool
	FoldInReplays bool
}
