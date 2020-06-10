package data_test

import (
	"GoSlice/data"
	"GoSlice/util/test"
	"github.com/google/go-cmp/cmp"
	"testing"
)

const (
	x = 10
	y = 20
	z = 30
)

// some helper functions

func setupMicroVec3() data.MicroVec3 {
	return data.NewMicroVec3(x, y, z)
}

func setupMicroPoint() data.MicroPoint {
	return data.NewMicroPoint(x, y)
}

// microVec3Comparer returns a cmp.Comparer which can handle MicroVec3.
func microVec3Comparer() cmp.Option {
	return cmp.Comparer(func(vec1, vec2 data.MicroVec3) bool {
		return vec1.X() == vec2.X() && vec1.Y() == vec2.Y() && vec1.Z() == vec2.Z()
	})
}

// microVec3Comparer returns a cmp.Comparer which can handle MicroPoint.
func microPointComparer() cmp.Option {
	return cmp.Comparer(func(vec1, vec2 data.MicroPoint) bool {
		return vec1.X() == vec2.X() && vec1.Y() == vec2.Y()
	})
}

// assertMicroVec3 checks if the vector vec contains the given xyz values.
func assertMicroVec3(t testing.TB, vec data.MicroVec3, xyz ...data.Micrometer) {
	if len(xyz) != 3 {
		// if it goes here, assertMicroVec3 is used wrong
		t.FailNow()
	}

	test.Assert(t, vec.X() == xyz[0], "X() should be %v but it is %v", x, vec.X())
	test.Assert(t, vec.Y() == xyz[1], "Y() should be %v but it is %v", y, vec.Y())
	test.Assert(t, vec.Z() == xyz[2], "Z() should be %v but it is %v", z, vec.Z())
}

// assertMicroPoint checks if the vector vec contains the given xy values.
func assertMicroPoint(t testing.TB, vec data.MicroPoint, xy ...data.Micrometer) {
	if len(xy) != 2 {
		// if it goes here, assertMicroVec3 is used wrong
		t.FailNow()
	}

	test.Assert(t, vec.X() == xy[0], "X() should be %v but it is %v", x, vec.X())
	test.Assert(t, vec.Y() == xy[1], "Y() should be %v but it is %v", y, vec.Y())
}

func TestMax(t *testing.T) {
	var tests = []struct {
		expected data.Micrometer
		a, b     data.Micrometer
	}{
		{3, 1, 3},
		{3, 3, 1},
		{-20, -30, -20},
		{-20, -20, -20},
	}

	for _, testCase := range tests {
		actual := data.Max(testCase.a, testCase.b)
		test.Equals(t, testCase.expected, actual)
	}
}

// Test MicroVec3 implementation

func TestNewMicroVec3(t *testing.T) {
	vec := data.NewMicroVec3(x, y, z)
	test.Assert(t, vec != nil, "vec should not be nil")

	assertMicroVec3(t, vec, x, y, z)
}

func TestMicroVec3Add(t *testing.T) {
	var expected = []data.Micrometer{20, 40, 60}

	vec := setupMicroVec3()
	vec2 := setupMicroVec3()
	actual := vec.Add(vec2)

	newInstance := setupMicroVec3()

	test.Assert(t, cmp.Equal(vec, newInstance, microVec3Comparer()), "the instance should not have been modified")
	assertMicroVec3(t, actual, expected...)
}

func TestMicroVec3Sub(t *testing.T) {
	var expected = []data.Micrometer{0, 0, 0}

	vec := setupMicroVec3()
	vec2 := setupMicroVec3()
	actual := vec.Sub(vec2)

	newInstance := setupMicroVec3()

	test.Assert(t, cmp.Equal(vec, newInstance, microVec3Comparer()), "the instance should not have been modified")
	assertMicroVec3(t, actual, expected...)
}

func TestMicroVec3Mul(t *testing.T) {
	var expected = []data.Micrometer{30, 60, 90}

	vec := setupMicroVec3()
	actual := vec.Mul(3)

	newInstance := setupMicroVec3()

	test.Assert(t, cmp.Equal(vec, newInstance, microVec3Comparer()), "the instance should not have been modified")
	assertMicroVec3(t, actual, expected...)
}

func TestMicroVec3Div(t *testing.T) {
	var expected = []data.Micrometer{5, 10, 15}

	vec := setupMicroVec3()
	actual := vec.Div(2)

	newInstance := setupMicroVec3()

	test.Assert(t, cmp.Equal(vec, newInstance, microVec3Comparer()), "the instance should not have been modified")
	assertMicroVec3(t, actual, expected...)
}

func TestMicroVec3Max(t *testing.T) {
	var tests = []struct {
		expected data.Micrometer
		vector   data.MicroVec3
	}{
		{3, data.NewMicroVec3(1, 2, 3)},
		{3, data.NewMicroVec3(1, 3, 2)},
		{3, data.NewMicroVec3(3, 2, 1)},

		{1, data.NewMicroVec3(1, 1, 1)},
		{1, data.NewMicroVec3(0, 0, 1)},
		{-5, data.NewMicroVec3(-10, -5, -10)},
	}

	for _, testCase := range tests {
		actual := testCase.vector.Max()
		test.Equals(t, testCase.expected, actual)
	}
}

func TestMicroVec3PointXY(t *testing.T) {
	vec := setupMicroVec3()
	result := vec.PointXY()

	assertMicroPoint(t, result, x, y)
}

func TestMicroVec3TestLength(t *testing.T) {
	vec := setupMicroVec3()

	var tests = []struct {
		expected bool
		length   data.Micrometer
	}{
		{true, 100},
		{true, 38},
		{false, 37},
		{false, 36},
		{false, 0},
	}

	for _, testCase := range tests {
		actual := vec.ShorterThanOrEqual(testCase.length)
		test.Assert(t, testCase.expected == actual, "the length %v should return %v but it returns %v", testCase.length, testCase.expected, actual)
	}
}

func TestMicroVec3TestSize2(t *testing.T) {
	var expected = data.Micrometer(1400)
	vec := setupMicroVec3()
	test.Equals(t, expected, vec.Size2())
}

func TestMicroVec3TestSize(t *testing.T) {
	var expected = data.Micrometer(37)
	vec := setupMicroVec3()
	test.Equals(t, expected, vec.Size())
}

func TestMicroVec3TestCopy(t *testing.T) {
	vec := setupMicroVec3()
	copied := vec.Copy()

	test.Assert(t, &copied != &vec, "Copy should create a new instance")
	test.Equals(t, vec, copied, microVec3Comparer())
}

func TestMicroVec3TestString(t *testing.T) {
	var expected = "10_20_30"
	vec := setupMicroVec3()

	test.Equals(t, expected, vec.String())
}

func TestMicroVec3TestSet(t *testing.T) {
	var expected = []data.Micrometer{40, 60, 200}

	actual := setupMicroVec3()
	err := actual.Set("40_60_200")

	test.Ok(t, err)
	assertMicroVec3(t, actual, expected...)
}

func TestMicroVec3TestType(t *testing.T) {
	var expected = "Micrometer"
	vec := setupMicroVec3()

	test.Equals(t, expected, vec.Type())
}

func TestMicroVec3TestSetXYZ(t *testing.T) {
	var expected = []data.Micrometer{50, 90, 200}
	actual := setupMicroVec3()

	actual.SetX(expected[0])
	actual.SetY(expected[1])
	actual.SetZ(expected[2])

	assertMicroVec3(t, actual, expected...)
}

// Test MicroPoint implementation

func TestNewMicroPoint(t *testing.T) {
	vec := data.NewMicroPoint(x, y)
	test.Assert(t, vec != nil, "vec should not be nil")

	assertMicroPoint(t, vec, x, y)
}

func TestMicroPointAdd(t *testing.T) {
	var expected = []data.Micrometer{20, 40}

	vec := setupMicroPoint()
	vec2 := setupMicroPoint()
	actual := vec.Add(vec2)

	newInstance := setupMicroPoint()
	test.Assert(t, cmp.Equal(vec, newInstance, microPointComparer()), "the instance should not have been modified")
	assertMicroPoint(t, actual, expected...)
}

func TestMicroPointSub(t *testing.T) {
	var expected = []data.Micrometer{0, 0}

	vec := setupMicroPoint()
	vec2 := setupMicroPoint()
	actual := vec.Sub(vec2)

	newInstance := setupMicroPoint()
	test.Assert(t, cmp.Equal(vec, newInstance, microPointComparer()), "the instance should not have been modified")
	assertMicroPoint(t, actual, expected...)
}

func TestMicroPointMul(t *testing.T) {
	var expected = []data.Micrometer{30, 60}

	vec := setupMicroPoint()
	actual := vec.Mul(3)

	newInstance := setupMicroPoint()
	test.Assert(t, cmp.Equal(vec, newInstance, microPointComparer()), "the instance should not have been modified")
	assertMicroPoint(t, actual, expected...)
}

func TestMicroPointDiv(t *testing.T) {
	var expected = []data.Micrometer{5, 10}

	vec := setupMicroPoint()
	actual := vec.Div(2)

	newInstance := setupMicroPoint()
	test.Assert(t, cmp.Equal(vec, newInstance, microPointComparer()), "the instance should not have been modified")
	assertMicroPoint(t, actual, expected...)
}

func TestMicroPointTestSize2(t *testing.T) {
	var expected = data.Micrometer(500)
	vec := setupMicroPoint()

	test.Equals(t, expected, vec.Size2())
}

func TestMicroPointTestSize(t *testing.T) {
	var expected = data.Micrometer(22)
	vec := setupMicroPoint()

	test.Equals(t, expected, vec.Size())
}

func TestMicroPointTestSizeMM(t *testing.T) {
	var expected = data.Millimeter(0.022360679)
	vec := setupMicroPoint()

	test.Equals(t, expected, vec.SizeMM())
}

func TestMicroPointTestCopy(t *testing.T) {
	vec := setupMicroPoint()

	copied := vec.Copy()

	test.Assert(t, &copied != &vec, "Copy should create a new instance")
	test.Equals(t, vec, copied, microPointComparer())
}

func TestMicroPointTestSetXY(t *testing.T) {
	var expected = []data.Micrometer{50, 90}
	actual := setupMicroPoint()

	actual.SetX(expected[0])
	actual.SetY(expected[1])

	assertMicroPoint(t, actual, expected...)
}
