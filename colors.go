package charts

var (
	Category10 Palette
	Tableau10  Palette
)

func init() {
	Category10 = splitColorString("1f77b4ff7f0e2ca02cd627289467bd8c564be377c27f7f7fbcbd2217becf")
	Tableau10 = splitColorString("4e79a7f28e2ce1575976b7b259a14fedc949af7aa1ff9da79c755fbab0ab")
}

func splitColorString(str string) []string {
	var arr []string
	for i := 0; i < len(str); i += 6 {
		arr = append(arr, "#"+str[i:i+6])
	}
	return arr
}
