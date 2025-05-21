package spatial

type directionValue int

const (
	left  = -1
	right = 1
)

type Direction struct {
	Value directionValue
}

func NewDirectionRight() Direction {
	return Direction{
		Value: right,
	}
}

func NewDirectionLeft() Direction {
	return Direction{
		Value: left,
	}
}

func (d *Direction) SetLeft() {
	d.Value = left
}

func (d *Direction) SetRight() {
	d.Value = right
}

func (d *Direction) IsRight() bool {
	return d.Value == right
}

func (d *Direction) IsLeft() bool {
	return d.Value == left
}

func (d *Direction) AsFloat() float64 {
	if d.Value == left {
		return -1
	} else {
		return 1
	}
}

func (d *Direction) Valid() bool {
	return d.Value == left || d.Value == right
}
