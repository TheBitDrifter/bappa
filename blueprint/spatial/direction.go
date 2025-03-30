package blueprintspatial

type directionValue int

const (
	left  = 0
	right = 1
)

type Direction struct {
	value directionValue
}

func NewDirectionRight() Direction {
	return Direction{
		value: right,
	}
}

func NewDirectionLeft() Direction {
	return Direction{
		value: left,
	}
}

func (d *Direction) SetLeft() {
	d.value = left
}

func (d *Direction) SetRight() {
	d.value = right
}

func (d *Direction) IsRight() bool {
	return d.value == right
}

func (d *Direction) IsLeft() bool {
	return d.value == left
}

func (d *Direction) AsFloat() float64 {
	if d.value == left {
		return -1
	} else {
		return 1
	}
}
