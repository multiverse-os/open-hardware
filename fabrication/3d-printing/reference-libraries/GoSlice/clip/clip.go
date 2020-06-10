// Package clip provides everything for calculating and altering polygons.
// The external clipper lib should only be used inside of this package.
package clip

import (
	"GoSlice/data"

	clipper "github.com/aligator/go.clipper"
)

// Pattern is an interface for all infill types which can be used to fill layer parts.
type Pattern interface {
	// Fill fills the given part.
	// It returns the final infill pattern.
	Fill(layerNr int, part data.LayerPart) data.Paths
}

// Clipper is an interface that provides methods needed by GoSlice to clip and alter polygons.
type Clipper interface {
	// GenerateLayerParts partitions the whole layer into several partition parts.
	// Each of them describes a polygon with holes.
	GenerateLayerParts(l data.Layer) (data.PartitionedLayer, bool)

	// InsetLayer returns all new paths generated by insetting all parts of the layer.
	// The result is built the following way: [part][insetNr][insetParts]data.LayerPart
	//
	//  * Part is the part in the from the input-layer.
	//  * Wall is the wall of the part. The first wall is the outer perimeter.
	//  * InsetNum is the number of the inset (starting by the outer walls with 0)
	//    and all following are from holes inside of the polygon.
	// The array for a part may be empty.
	//
	// If you need to ex-set a part, just provide a negative offset.
	InsetLayer(layer []data.LayerPart, offset data.Micrometer, insetCount int) [][][]data.LayerPart

	// Inset insets the given layer part.
	// The result is built the following way: [insetNr][insetParts]data.LayerPart
	//
	//  * Wall is the wall of the part. The first wall is the outer perimeter
	//  * InsetNum is the number of the inset (starting by the outer walls with 0)
	//    and all following are from holes inside of the polygon.
	// The array for a part may be empty.
	//
	// If you need to ex-set a part, just provide a negative offset.
	Inset(part data.LayerPart, offset data.Micrometer, insetCount int) [][]data.LayerPart

	// Difference calculates the difference between the parts and the toRemove parts.
	// It returns the result as a new slice of layer parts.
	Difference(parts []data.LayerPart, toRemove []data.LayerPart) (clippedParts []data.LayerPart, ok bool)

	// Intersection calculates the intersection between the parts and the toIntersect parts.
	// It returns the result as a new slice of layer parts.
	Intersection(parts []data.LayerPart, toIntersect []data.LayerPart) (clippedParts []data.LayerPart, ok bool)

	// Union calculates the union of the parts and the toMerge parts.
	// It returns the result as a new slice of layer parts.
	Union(parts []data.LayerPart, toIntersect []data.LayerPart) (clippedParts []data.LayerPart, ok bool)
}

// clipperClipper implements Clipper using the external clipper library.
type clipperClipper struct{}

// NewClipper returns a new instance of a polygon Clipper.
func NewClipper() Clipper {
	return &clipperClipper{}
}

// clipperPoint converts the GoSlice point representation to the
// representation which is used by the external clipper lib.
func clipperPoint(p data.MicroPoint) *clipper.IntPoint {
	return &clipper.IntPoint{
		X: clipper.CInt(p.X()),
		Y: clipper.CInt(p.Y()),
	}
}

// clipperPaths converts the GoSlice Paths representation
// to the representation which is used by the external clipper lib.
func clipperPaths(p data.Paths) clipper.Paths {
	var result clipper.Paths
	for _, path := range p {
		result = append(result, clipperPath(path))
	}

	return result
}

// clipperPath converts the GoSlice Path representation
// to the representation which is used by the external clipper lib.
func clipperPath(p data.Path) clipper.Path {
	var result clipper.Path
	for _, point := range p {
		result = append(result, clipperPoint(point))
	}

	return result
}

// microPoint converts the external clipper lib representation of a point
// to the representation which is used by GoSlice.
func microPoint(p *clipper.IntPoint) data.MicroPoint {
	return data.NewMicroPoint(data.Micrometer(p.X), data.Micrometer(p.Y))
}

// microPath converts the external clipper lib representation of a path
// to the representation which is used by GoSlice.
// The parameter simplify enables simplifying of the path using
// the default simplification settings.
func microPath(p clipper.Path, simplify bool) data.Path {
	var result data.Path
	for _, point := range p {
		result = append(result, microPoint(point))
	}

	if simplify {
		return result.Simplify(-1, -1)
	}
	return result
}

// microPaths converts the external clipper lib representation of paths
// to the representation which is used by GoSlice.
// The parameter simplify enables simplifying of the paths using
// the default simplification settings.
func microPaths(p clipper.Paths, simplify bool) data.Paths {
	var result data.Paths
	for _, path := range p {
		result = append(result, microPath(path, simplify))
	}
	return result
}

func (c clipperClipper) GenerateLayerParts(l data.Layer) (data.PartitionedLayer, bool) {
	polyList := clipper.Paths{}
	// convert all polygons to clipper polygons
	for _, layerPolygon := range l.Polygons() {
		polyList = append(polyList, clipperPath(layerPolygon.Simplify(-1, -1)))
	}

	if len(polyList) == 0 {
		return data.NewPartitionedLayer([]data.LayerPart{}), true
	}

	cl := clipper.NewClipper(clipper.IoNone)
	cl.AddPaths(polyList, clipper.PtSubject, true)
	resultPolys, ok := cl.Execute2(clipper.CtUnion, clipper.PftEvenOdd, clipper.PftEvenOdd)
	if !ok {
		return nil, false
	}

	return data.NewPartitionedLayer(polyTreeToLayerParts(resultPolys)), true
}

// polyTreeToLayerParts creates layer parts out of a poly tree (which is the result of clipper's Execute2).
func polyTreeToLayerParts(tree *clipper.PolyTree) []data.LayerPart {
	var layerParts []data.LayerPart

	var polysForNextRound []*clipper.PolyNode

	for _, c := range tree.Childs() {
		polysForNextRound = append(polysForNextRound, c)
	}
	for {
		if polysForNextRound == nil {
			break
		}
		thisRound := polysForNextRound
		polysForNextRound = nil

		for _, p := range thisRound {
			var holes data.Paths

			for _, child := range p.Childs() {
				// TODO: simplify, yes / no ??
				holes = append(holes, microPath(child.Contour(), false))
				for _, c := range child.Childs() {
					polysForNextRound = append(polysForNextRound, c)
				}
			}

			// TODO: simplify, yes / no ??
			layerParts = append(layerParts, data.NewBasicLayerPart(microPath(p.Contour(), false), holes))
		}
	}

	return layerParts
}

func (c clipperClipper) InsetLayer(layer []data.LayerPart, offset data.Micrometer, insetCount int) [][][]data.LayerPart {
	var result [][][]data.LayerPart
	for _, part := range layer {
		result = append(result, c.Inset(part, offset, insetCount))
	}

	return result
}

func (c clipperClipper) Inset(part data.LayerPart, offset data.Micrometer, insetCount int) [][]data.LayerPart {
	var insets [][]data.LayerPart

	co := clipper.NewClipperOffset()

	for insetNr := 0; insetNr < insetCount; insetNr++ {
		// insets for the outline
		co.Clear()
		co.AddPaths(clipperPaths(data.Paths{part.Outline()}), clipper.JtSquare, clipper.EtClosedPolygon)
		co.AddPaths(clipperPaths(part.Holes()), clipper.JtSquare, clipper.EtClosedPolygon)

		co.MiterLimit = 2
		allNewInsets := co.Execute2(float64(-int(offset)*insetNr) - float64(offset/2))
		insets = append(insets, polyTreeToLayerParts(allNewInsets))
	}

	return insets
}

func (c clipperClipper) Difference(parts []data.LayerPart, toRemove []data.LayerPart) (clippedParts []data.LayerPart, ok bool) {
	return c.runClipper(clipper.CtDifference, parts, toRemove)
}

func (c clipperClipper) Intersection(parts []data.LayerPart, toIntersect []data.LayerPart) (clippedParts []data.LayerPart, ok bool) {
	return c.runClipper(clipper.CtIntersection, parts, toIntersect)
}

func (c clipperClipper) Union(parts []data.LayerPart, toMerge []data.LayerPart) (clippedParts []data.LayerPart, ok bool) {
	return c.runClipper(clipper.CtUnion, parts, toMerge)
}

func (c clipperClipper) runClipper(clipType clipper.ClipType, parts []data.LayerPart, toClip []data.LayerPart) (clippedParts []data.LayerPart, ok bool) {
	cl := clipper.NewClipper(clipper.IoNone)
	for _, part := range parts {
		cl.AddPath(clipperPath(part.Outline()), clipper.PtSubject, true)
		cl.AddPaths(clipperPaths(part.Holes()), clipper.PtSubject, true)
	}

	for _, intersect := range toClip {
		cl.AddPath(clipperPath(intersect.Outline()), clipper.PtClip, true)
		cl.AddPaths(clipperPaths(intersect.Holes()), clipper.PtClip, true)
	}

	tree, ok := cl.Execute2(clipType, clipper.PftEvenOdd, clipper.PftEvenOdd)

	if !ok {
		return nil, ok
	}
	return polyTreeToLayerParts(tree), ok
}
