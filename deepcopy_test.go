package deepcopy

import (
	"fmt"
	. "reflect"
	"testing"
	"time"
)

func ExampleAnything() {
	tests := []interface{}{
		`"Now cut that out!"`,
		39,
		true,
		false,
		2.14,
		[]string{
			"Phil Harris",
			"Rochester van Jones",
			"Mary Livingstone",
			"Dennis Day",
		},
		[2]string{
			"Jell-O",
			"Grape-Nuts",
		},
	}

	for _, expected := range tests {
		actual := MustAnything(expected)
		fmt.Println(actual)
	}
	// Output:
	// "Now cut that out!"
	// 39
	// true
	// false
	// 2.14
	// [Phil Harris Rochester van Jones Mary Livingstone Dennis Day]
	// [Jell-O Grape-Nuts]
}

func ExampleMap() {

	type Inner struct {
		Value int
	}

	type Foo struct {
		Inner *Inner
		Bar   int
	}

	x := map[string]*Foo{
		"elem1": &Foo{
			Inner: &Inner{Value: 3},
			Bar:   1,
		},
		"elem2": &Foo{
			Inner: nil,
			Bar:   2,
		},
	}

	y := MustAnything(x).(map[string]*Foo)

	fmt.Printf("x[elem1] = y[elem1]: %v\n", x["elem1"] == y["elem1"])
	fmt.Printf("x[elem1].Inner = y[elem1].Inner: %v\n", x["elem1"].Inner == y["elem1"].Inner)
	fmt.Printf("x[elem1].Inner = y[elem1].Inner: %v\n", *x["elem1"].Inner == *y["elem1"].Inner)
	fmt.Printf("x[elem1].Bar = y[elem1].Bar: %v\n", x["elem1"].Bar == y["elem1"].Bar)

	fmt.Printf("x[elem2] = y[elem2]: %v\n", x["elem2"] == y["elem2"])
	fmt.Printf("x[elem2].Inner = y[elem2].Inner: %v\n", x["elem2"].Inner == y["elem2"].Inner)
	fmt.Printf("x[elem2].Inner = nil: %v\n", x["elem2"].Inner == nil)
	fmt.Printf("x[elem2].Bar = y[elem2].Bar: %v\n", x["elem2"].Bar == y["elem2"].Bar)

	// Output:
	// x[elem1] = y[elem1]: false
	// x[elem1].Inner = y[elem1].Inner: false
	// x[elem1].Inner = y[elem1].Inner: true
	// x[elem1].Bar = y[elem1].Bar: true
	// x[elem2] = y[elem2]: false
	// x[elem2].Inner = y[elem2].Inner: true
	// x[elem2].Inner = nil: true
	// x[elem2].Bar = y[elem2].Bar: true
}

func TestInterface(t *testing.T) {
	x := []interface{}{nil}
	y := MustAnything(x).([]interface{})
	if !DeepEqual(x, y) || len(y) != 1 {
		t.Errorf("expect %v == %v; y had length %v (expected 1)", x, y, len(y))
	}
	var a interface{}
	b := MustAnything(a)
	if a != b {
		t.Errorf("expected %v == %v", a, b)
	}
}

func ExampleAvoidInfiniteLoops() {

	type Foo struct {
		Foo *Foo
		Bar int
	}

	x := &Foo{
		Bar: 4,
	}
	x.Foo = x
	y := MustAnything(x).(*Foo)
	fmt.Printf("x == y: %v\n", x == y)
	fmt.Printf("x == x.Foo: %v\n", x == x.Foo)
	fmt.Printf("y == y.Foo: %v\n", y == y.Foo)
	// Output:
	// x == y: false
	// x == x.Foo: true
	// y == y.Foo: true
}

func TestUnsupportedKind(t *testing.T) {
	x := func() {}

	tests := []interface{}{
		x,
		map[bool]interface{}{true: x},
		[]interface{}{x},
	}

	for _, test := range tests {
		y, err := Anything(test)
		if y != nil {
			t.Errorf("expected %v to be nil", y)
		}
		if err == nil {
			t.Errorf("expected err to not be nil")
		}
	}
}

func TestUnsupportedKindPanicsOnMust(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected a panic; didn't get one")
		}
	}()
	x := func() {}
	MustAnything(x)
}

func TestMismatchedTypesFail(t *testing.T) {
	tests := []struct {
		input interface{}
		kind  Kind
	}{
		{
			map[int]int{1: 2, 2: 4, 3: 8},
			Map,
		},
		{
			[]int{2, 8},
			Slice,
		},
	}
	for _, test := range tests {
		for kind, copier := range copiers {
			if kind == test.kind {
				continue
			}
			actual, err := copier(test.input, nil)
			if actual != nil {

				t.Errorf("%v attempted value %v as %v; should be nil value, got %v", test.kind, test.input, kind, actual)
			}
			if err == nil {
				t.Errorf("%v attempted value %v as %v; should have gotten an error", test.kind, test.input, kind)
			}
		}
	}
}

func TestNilMaps(t *testing.T) {
	type Foo struct {
		A int
	}

	type Bar struct {
		A  map[string]Foo
		Aa map[string]Foo
		B  map[string]*Foo
		Bb map[string]*Foo
		C  *map[string]*Foo
		Cc *map[string]*Foo
		D  map[*Foo]int
	}

	src := Bar{
		A: map[string]Foo{
			"t1": {A: 7},
			"t2": {A: 8},
		},
		Aa: nil,
		B: map[string]*Foo{
			"t1": {A: 7},
			"t2": nil,
		},
		Bb: nil,
		C: &map[string]*Foo{
			"t1": nil,
		},
		Cc: nil,
		D: map[*Foo]int{
			{A: 9}: 3,
			nil:    6,
		},
	}

	dst := MustAnything(src)
	dstTyped, ok := dst.(Bar)
	if !ok {
		t.Errorf("failed to convert to concrete type after deepCopy")
	}

	if src.C == dstTyped.C {
		t.Errorf("pointers are equal, expect %p != %p; ", src.C, dstTyped.C)
	}
	if src.Cc != nil || dstTyped.Cc != nil {
		t.Errorf("pointers are not nil, expect %p == nil; %p == nil; ", src.Cc, dstTyped.Cc)
	}
	if src.B["t1"] == dstTyped.B["t1"] {
		t.Errorf("pointers are equal, expect %p != %p; ", src.C, dstTyped.C)
	}
	if src.B["t2"] != nil || dstTyped.B["t2"] != nil {
		t.Errorf("pointers are not nil, expect %p == nil; %p == nil; ", src.B["t2"], dstTyped.B["t2"])
	}
	if src.Bb != nil || dstTyped.Bb != nil {
		t.Errorf("pointers are not nil, expect %p == nil; %p == nil; ", src.Bb, dstTyped.Bb)
	}

	if !DeepEqual(src, dstTyped) {
		t.Errorf("expect %v == %v", src, dstTyped)
	}
}

func TestNilSlice(t *testing.T) {
	type Foo struct {
		A []string
		B []string
	}

	src := Foo{
		A: []string{"t1", "t2"},
		B: nil,
	}
	dst := MustAnything(src)

	if !DeepEqual(src, dst) {
		t.Errorf("expect %v == %v; ", src, dst)
	}
}

func TestNilPointer(t *testing.T) {
	type Foo struct {
		A int
	}
	type Bar struct {
		B int
	}
	type FooBar struct {
		Foo  *Foo
		Bar  *Bar
		Foo2 *Foo
		Bar2 *Bar
	}

	src := &FooBar{
		Foo2: &Foo{1},
		Bar2: &Bar{2},
	}

	dst := MustAnything(src)
	dstTyped, ok := dst.(*FooBar)
	if !ok {
		t.Errorf("failed to convert to concrete type after deepCopy")
	}

	if src.Foo != nil || dstTyped.Foo != nil {
		t.Errorf("pointers are not nil, expect %p == nil; %p == nil; ", src.Foo, dstTyped.Foo)
	}
	if src.Bar != nil || dstTyped.Bar != nil {
		t.Errorf("pointers are not nil, expect %p == nil; %p == nil; ", src.Bar, dstTyped.Bar)
	}
	if src.Foo2 == dstTyped.Foo2 {
		t.Errorf("pointers are equal, expect %p != %p; ", src.Foo2, dstTyped.Foo2)
	}
	if src.Bar2 == dstTyped.Bar2 {
		t.Errorf("pointers are equal, expect %p != %p; ", src.Bar2, dstTyped.Bar2)
	}

	if !DeepEqual(src, dst) {
		t.Errorf("expect %v == %v; ", src, dst)
	}

}

func TestStructTime(t *testing.T) {
	type Foo struct {
		Time1 time.Time
		Time2 time.Time
	}

	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}

	src := Foo{
		Time1: time.Now().In(location),
		Time2: time.Now().UTC(),
	}
	dst := MustAnything(src)

	dstTyped, ok := dst.(Foo)
	if !ok {
		t.Errorf("failed to convert to concrete type after deepCopy")
	}

	if src.Time1.Format(time.RFC3339Nano) != dstTyped.Time1.Format(time.RFC3339Nano) {
		t.Errorf("Time1 is not equal, expect %v == %v; ", src.Time1, dstTyped.Time1)
	}
	if !src.Time1.Equal(dstTyped.Time1) {
		t.Errorf("Time1 is not equal, expect %v == %v; ", src.Time1, dstTyped.Time1)
	}

	if src.Time2.Format(time.RFC3339Nano) != dstTyped.Time2.Format(time.RFC3339Nano) {
		t.Errorf("Time2 is not equal, expect %v == %v; ", src.Time2, dstTyped.Time1)
	}
	if !src.Time2.Equal(dstTyped.Time2) {
		t.Errorf("Time2 is not equal, expect %v == %v; ", src.Time2, dstTyped.Time2)
	}

	if !DeepEqual(src, dst) {
		t.Errorf("expect %v ==", src)
		t.Errorf("    == %v; ", dst)
	}
}

func TestSliceTime(t *testing.T) {

	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}

	src := make([]time.Time, 2)
	src[0] = time.Now().In(location)
	src[1] = time.Now().UTC()

	dst := MustAnything(src)

	dstTyped, ok := dst.([]time.Time)
	if !ok {
		t.Errorf("failed to convert to concrete type after deepCopy")
	}

	if src[0].Format(time.RFC3339Nano) != dstTyped[0].Format(time.RFC3339Nano) {
		t.Errorf("src[0] is not equal, expect %v == %v; ", src[0], dstTyped[0])
	}
	if !src[0].Equal(dstTyped[0]) {
		t.Errorf("src[0] is not equal, expect %v == %v; ", src[0], dstTyped[0])
	}

	if src[1].Format(time.RFC3339Nano) != dstTyped[1].Format(time.RFC3339Nano) {
		t.Errorf("src[1] is not equal, expect %v == %v; ", src[1], dstTyped[1])
	}
	if !src[1].Equal(dstTyped[1]) {
		t.Errorf("src[1] is not equal, expect %v == %v; ", src[1], dstTyped[1])
	}

	if &src[0] == &dstTyped[0] {
		t.Errorf("pointers are equal, expect %p != %p; ", &src[0], &dstTyped[0])
	}

	if &src[1] == &dstTyped[1] {
		t.Errorf("pointers are equal, expect %p != %p; ", &src[1], &dstTyped[1])
	}

	if !DeepEqual(src, dst) {
		t.Errorf("expect %v ==", src)
		t.Errorf("    == %v; ", dst)
	}
}

func TestSliceTimePointer(t *testing.T) {

	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}

	src := make([]*time.Time, 2)
	temp := time.Now().In(location)
	src[0] = &temp
	temp = time.Now().UTC()
	src[1] = &temp

	dst := MustAnything(src)

	dstTyped, ok := dst.([]*time.Time)
	if !ok {
		t.Errorf("failed to convert to concrete type after deepCopy")
	}

	if src[0].Format(time.RFC3339Nano) != dstTyped[0].Format(time.RFC3339Nano) {
		t.Errorf("src[0] is not equal, expect %v == %v; ", src[0], dstTyped[0])
	}
	if !src[0].Equal(*dstTyped[0]) {
		t.Errorf("src[0] is not equal, expect %v == %v; ", src[0], dstTyped[0])
	}

	if src[1].Format(time.RFC3339Nano) != dstTyped[1].Format(time.RFC3339Nano) {
		t.Errorf("src[1] is not equal, expect %v == %v; ", src[1], dstTyped[1])
	}
	if !src[1].Equal(*dstTyped[1]) {
		t.Errorf("src[1] is not equal, expect %v == %v; ", src[1], dstTyped[1])
	}

	if src[0] == dstTyped[0] {
		t.Errorf("pointers are equal, expect %p != %p; ", &src[0], &dstTyped[0])
	}

	if src[1] == dstTyped[1] {
		t.Errorf("pointers are equal, expect %p != %p; ", &src[1], &dstTyped[1])
	}

	if !DeepEqual(src, dst) {
		t.Errorf("expect %v ==", src)
		t.Errorf("    == %v; ", dst)
	}
}

func TestMapTime(t *testing.T) {

	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}

	src := make(map[string]time.Time)
	src["t1"] = time.Now().In(location)
	src["t2"] = time.Now().UTC()

	dst := MustAnything(src)

	dstTyped, ok := dst.(map[string]time.Time)
	if !ok {
		t.Errorf("failed to convert to concrete type after deepCopy")
	}

	if src["t1"].Format(time.RFC3339Nano) != dstTyped["t1"].Format(time.RFC3339Nano) {
		t.Errorf("src[t1] is not equal, expect %v == %v; ", src["t1"], dstTyped["t1"])
	}
	if !src["t1"].Equal(dstTyped["t1"]) {
		t.Errorf("src[t1] is not equal, expect %v == %v; ", src["t1"], dstTyped["t1"])
	}

	if src["t2"].Format(time.RFC3339Nano) != dstTyped["t2"].Format(time.RFC3339Nano) {
		t.Errorf("src[t2] is not equal, expect %v == %v; ", src["t2"], dstTyped["t2"])
	}
	if !src["t2"].Equal(dstTyped["t2"]) {
		t.Errorf("src[t2] is not equal, expect %v == %v; ", src["t2"], dstTyped["t2"])
	}

	if !DeepEqual(src, dst) {
		t.Errorf("expect %v ==", src)
		t.Errorf("    == %v; ", dst)
	}
}

func TestMapTimePointer(t *testing.T) {

	location, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}

	src := make(map[string]*time.Time)
	temp := time.Now().In(location)
	src["t1"] = &temp
	temp = time.Now().UTC()
	src["t2"] = &temp

	dst := MustAnything(src)

	dstTyped, ok := dst.(map[string]*time.Time)
	if !ok {
		t.Errorf("failed to convert to concrete type after deepCopy")
	}

	if src["t1"].Format(time.RFC3339Nano) != dstTyped["t1"].Format(time.RFC3339Nano) {
		t.Errorf("src[t1] is not equal, expect %v == %v; ", src["t1"], dstTyped["t1"])
	}
	if !src["t1"].Equal(*dstTyped["t1"]) {
		t.Errorf("src[t1] is not equal, expect %v == %v; ", src["t1"], dstTyped["t1"])
	}

	if src["t2"].Format(time.RFC3339Nano) != dstTyped["t2"].Format(time.RFC3339Nano) {
		t.Errorf("src[t2] is not equal, expect %v == %v; ", src["t2"], dstTyped["t2"])
	}
	if !src["t2"].Equal(*dstTyped["t2"]) {
		t.Errorf("src[t2] is not equal, expect %v == %v; ", src["t2"], dstTyped["t2"])
	}

	if src["t1"] == dstTyped["t1"] {
		t.Errorf("pointers are equal, expect %p != %p; ", src["t1"], dstTyped["t1"])
	}

	if src["t2"] == dstTyped["t2"] {
		t.Errorf("pointers are equal, expect %p != %p; ", src["t2"], dstTyped["t2"])
	}

	if !DeepEqual(src, dst) {
		t.Errorf("expect %v ==", src)
		t.Errorf("    == %v; ", dst)
	}
}
