package valueobject

import (
	"github.com/disintegration/gift"
	"math"
	"strings"
)

var imageFilters = map[string]gift.Resampling{
	strings.ToLower("NearestNeighbor"):   gift.NearestNeighborResampling,
	strings.ToLower("Box"):               gift.BoxResampling,
	strings.ToLower("Linear"):            gift.LinearResampling,
	strings.ToLower("Hermite"):           hermiteResampling,
	strings.ToLower("MitchellNetravali"): mitchellNetravaliResampling,
	strings.ToLower("CatmullRom"):        catmullRomResampling,
	strings.ToLower("BSpline"):           bSplineResampling,
	strings.ToLower("Gaussian"):          gaussianResampling,
	strings.ToLower("Lanczos"):           gift.LanczosResampling,
	strings.ToLower("Hann"):              hannResampling,
	strings.ToLower("Hamming"):           hammingResampling,
	strings.ToLower("Blackman"):          blackmanResampling,
	strings.ToLower("Bartlett"):          bartlettResampling,
	strings.ToLower("Welch"):             welchResampling,
	strings.ToLower("Cosine"):            cosineResampling,
}

var (
	// Hermite cubic spline filter (BC-spline; B=0; C=0).
	hermiteResampling = resamp{
		name:    "Hermite",
		support: 1.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 1.0 {
				return bcspline(x, 0.0, 0.0)
			}
			return 0
		},
	}

	// Mitchell-Netravali cubic filter (BC-spline; B=1/3; C=1/3).
	mitchellNetravaliResampling = resamp{
		name:    "MitchellNetravali",
		support: 2.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 2.0 {
				return bcspline(x, 1.0/3.0, 1.0/3.0)
			}
			return 0
		},
	}

	// Catmull-Rom - sharp cubic filter (BC-spline; B=0; C=0.5).
	catmullRomResampling = resamp{
		name:    "CatmullRomResampling",
		support: 2.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 2.0 {
				return bcspline(x, 0.0, 0.5)
			}
			return 0
		},
	}

	// BSpline is a smooth cubic filter (BC-spline; B=1; C=0).
	bSplineResampling = resamp{
		name:    "BSplineResampling",
		support: 2.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 2.0 {
				return bcspline(x, 1.0, 0.0)
			}
			return 0
		},
	}

	// Gaussian blurring filter.
	gaussianResampling = resamp{
		name:    "GaussianResampling",
		support: 2.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 2.0 {
				return float32(math.Exp(float64(-2 * x * x)))
			}
			return 0
		},
	}

	//  Hann-windowed sinc filter (3 lobes).
	hannResampling = resamp{
		name:    "HannResampling",
		support: 3.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 3.0 {
				return sinc(x) * float32(0.5+0.5*math.Cos(math.Pi*float64(x)/3.0))
			}
			return 0
		},
	}

	hammingResampling = resamp{
		name:    "HammingResampling",
		support: 3.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 3.0 {
				return sinc(x) * float32(0.54+0.46*math.Cos(math.Pi*float64(x)/3.0))
			}
			return 0
		},
	}

	// Blackman-windowed sinc filter (3 lobes).
	blackmanResampling = resamp{
		name:    "BlackmanResampling",
		support: 3.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 3.0 {
				return sinc(x) * float32(0.42-0.5*math.Cos(math.Pi*float64(x)/3.0+math.Pi)+0.08*math.Cos(2.0*math.Pi*float64(x)/3.0))
			}
			return 0
		},
	}

	bartlettResampling = resamp{
		name:    "BartlettResampling",
		support: 3.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 3.0 {
				return sinc(x) * (3.0 - x) / 3.0
			}
			return 0
		},
	}

	// Welch-windowed sinc filter (parabolic window, 3 lobes).
	welchResampling = resamp{
		name:    "WelchResampling",
		support: 3.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 3.0 {
				return sinc(x) * (1.0 - (x * x / 9.0))
			}
			return 0
		},
	}

	// Cosine-windowed sinc filter (3 lobes).
	cosineResampling = resamp{
		name:    "CosineResampling",
		support: 3.0,
		kernel: func(x float32) float32 {
			x = absf32(x)
			if x < 3.0 {
				return sinc(x) * float32(math.Cos((math.Pi/2.0)*(float64(x)/3.0)))
			}
			return 0
		},
	}
)

// The following code is borrowed from https://raw.githubusercontent.com/disintegration/gift/master/resize.go
// MIT licensed.
type resamp struct {
	name    string
	support float32
	kernel  func(float32) float32
}

func (r resamp) String() string {
	return r.name
}

func (r resamp) Support() float32 {
	return r.support
}

func (r resamp) Kernel(x float32) float32 {
	return r.kernel(x)
}

func bcspline(x, b, c float32) float32 {
	if x < 0 {
		x = -x
	}
	if x < 1 {
		return ((12-9*b-6*c)*x*x*x + (-18+12*b+6*c)*x*x + (6 - 2*b)) / 6
	}
	if x < 2 {
		return ((-b-6*c)*x*x*x + (6*b+30*c)*x*x + (-12*b-48*c)*x + (8*b + 24*c)) / 6
	}
	return 0
}

func absf32(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func sinc(x float32) float32 {
	if x == 0 {
		return 1
	}
	return float32(math.Sin(math.Pi*float64(x)) / (math.Pi * float64(x)))
}
