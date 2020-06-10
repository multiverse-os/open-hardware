package vip

import (
	"github.com/l1va/goosli/debug"
	"github.com/l1va/goosli/gcode"
	"github.com/l1va/goosli/helpers"
	. "github.com/l1va/goosli/primitives"
	. "github.com/l1va/goosli/slicers"
	"log"
)

func SliceRotation(mesh *Mesh, settings Settings, layers []Layer) gcode.Gcode {

	print("ROTATION slicing:\n")

	debug.RecreateFile()
	gcd := gcode.NewGcode(*settings.GcodeSettings)

	fillPlanes, fullFillPlanes := CalcFillPlanes(mesh, settings)

	i := 0
	for i < len(layers) {
		if i == 0 {
			i++
			continue
		}
		if failRotaion(layers[i-1], layers[i]) {
			break
		}
		i++
	}
	// Add horizontal layers
	gcd.AddLayers(PrepareLayers(layers[:i], settings, fillPlanes, fullFillPlanes))

	anyPoint := layers[i].Paths[0].Points[0]
	mesh, _, err := helpers.CutMesh(mesh, Plane{anyPoint, AxisZ})
	if err != nil {
		log.Fatal("failed to cut mesh, before rotation: ", err)
	}

	// Rotate bed
	mesh = mesh.Rotate(RotationAroundX(angleX), OriginPoint)
	gcd.Add(gcode.InclineX{})

	for i, plane := range fillPlanes {
		fillPlanes[i] = plane.Rotate(RotationAroundX(angleX))
	}
	for i, plane := range fullFillPlanes {
		fullFillPlanes[i] = plane.Rotate(RotationAroundX(angleX))
	}

	rest := SliceByVector(mesh, AxisZ.Rotate(RotationAroundX(angleX)), settings)
	gcd.AddLayers(PrepareLayers(rest, settings, fillPlanes, fullFillPlanes))

	gcd.Add(gcode.InclineXBack{})
	return gcd
}

func failRotaion(prevLayer, curLayer Layer) bool {

	if len(prevLayer.Paths) != 1 {
		println("prevLayer has not one path: ", len(prevLayer.Paths))
	}

	if len(curLayer.Paths) != 1 {
		println("curLayer has not one path: ", len(curLayer.Paths))
	}

	anyP := prevLayer.Paths[0].Points[0]
	//nearestP := curLayer.Paths[0].Lines[0].P1
	//d := anyP.DistanceTo(nearestP)
	//for _, line := range curLayer.Paths[0].Lines {
	//	d2 := line.P2.DistanceTo(anyP)
	//	if d2 < d {
	//		d = d2
	//		nearestP = line.P2
	//	}
	//}
	cp := FindCentroid(prevLayer.Paths[0])
	plane := Plane{anyP, cp.VectorTo(anyP).Cross(AxisZ)}
	nearestP := plane.IntersectPathCodirectedWith(curLayer.Paths[0], cp.VectorTo(anyP))

	if nearestP != nil && AxisZ.Angle(anyP.VectorTo(*nearestP)) > failAngle {
		if cp.DistanceTo(anyP) > cp.DistanceTo(*nearestP) { // inside incline
			println("inside incline: ", AxisZ.Angle(anyP.VectorTo(*nearestP)))
			return false
		}
		println("outside incline:", AxisZ.Angle(anyP.VectorTo(*nearestP)))
		debug.AddLine(Line{anyP, *nearestP}, debug.GreenColor)
		return true
	}
	return false

}
