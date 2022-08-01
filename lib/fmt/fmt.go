package fmt

//js-bind
//throw new Error("use fmt.Println instead")
func Print(args ...interface{}) {}

//js-bind
//console.log(%args%)
func Println(args ...interface{}) {}

//js-bind
//throw new Error("use fmt.Println instead")
func Printf(fmt string, args ...interface{}) {}
