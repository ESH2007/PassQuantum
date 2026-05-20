package palette

import (
	"image"
	"image/color"
	"math/rand"
	"sort"
	"time"
)

type ColorCluster struct {
	Color color.RGBA
	Count int
}

func SamplePixels(img image.Image, maxSamples int) []color.RGBA {
	if img == nil || maxSamples <= 0 {
		return nil
	}

	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()
	if width <= 0 || height <= 0 {
		return nil
	}

	total := width * height
	step := 1
	if total > maxSamples {
		step = (total + maxSamples - 1) / maxSamples
	}

	samples := make([]color.RGBA, 0, maxSamples)
	for idx := 0; idx < total; idx += step {
		x := b.Min.X + (idx % width)
		y := b.Min.Y + (idx / width)
		px := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
		if px.A < 128 {
			continue
		}
		samples = append(samples, px)
		if len(samples) >= maxSamples {
			break
		}
	}

	return samples
}

func KMeans(pixels []color.RGBA, k int, iterations int) []ColorCluster {
	if len(pixels) == 0 || k <= 0 || iterations <= 0 {
		return nil
	}
	if k > len(pixels) {
		k = len(pixels)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	centroids := make([]color.RGBA, k)
	perm := rng.Perm(len(pixels))
	for i := 0; i < k; i++ {
		centroids[i] = pixels[perm[i]]
	}

	assignments := make([]int, len(pixels))

	for it := 0; it < iterations; it++ {
		counts := make([]int, k)
		rSums := make([]int, k)
		gSums := make([]int, k)
		bSums := make([]int, k)

		for i, p := range pixels {
			closest := 0
			best := colorDistanceSquared(p, centroids[0])
			for j := 1; j < k; j++ {
				d := colorDistanceSquared(p, centroids[j])
				if d < best {
					best = d
					closest = j
				}
			}
			assignments[i] = closest
			counts[closest]++
			rSums[closest] += int(p.R)
			gSums[closest] += int(p.G)
			bSums[closest] += int(p.B)
		}

		for i := 0; i < k; i++ {
			if counts[i] == 0 {
				centroids[i] = pixels[rng.Intn(len(pixels))]
				continue
			}
			centroids[i] = color.RGBA{
				R: uint8(rSums[i] / counts[i]),
				G: uint8(gSums[i] / counts[i]),
				B: uint8(bSums[i] / counts[i]),
				A: 255,
			}
		}
	}

	counts := make([]int, k)
	for _, idx := range assignments {
		counts[idx]++
	}

	clusters := make([]ColorCluster, 0, k)
	for i := 0; i < k; i++ {
		clusters = append(clusters, ColorCluster{Color: centroids[i], Count: counts[i]})
	}

	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Count > clusters[j].Count
	})

	return clusters
}

func colorDistanceSquared(a color.RGBA, b color.RGBA) int {
	dr := int(a.R) - int(b.R)
	dg := int(a.G) - int(b.G)
	db := int(a.B) - int(b.B)
	return dr*dr + dg*dg + db*db
}
