package entity

type Color string

const (
	Blue      Color = "blue"
	Brown     Color = "brown"
	Gray      Color = "default"
	LightGray Color = "gray"
	Green     Color = "green"
	Orange    Color = "orange"
	Pink      Color = "pink"
	Purple    Color = "purple"
	Red       Color = "red"
	Yellow    Color = "yellow"
)

var Colors = []Color{
	Blue,
	Red,
	Green,
	Purple,
	Yellow,
	Pink,
	Orange,
	LightGray,
	Brown,
	Gray,
}

var ColorsMap = map[Color]struct{}{
	Blue:      {},
	Red:       {},
	Green:     {},
	Purple:    {},
	Yellow:    {},
	Pink:      {},
	Orange:    {},
	LightGray: {},
	Brown:     {},
	Gray:      {},
}
