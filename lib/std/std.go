package std

//js-bind
//%arg0%.toString()
func Stringify(subject any) string {}

//js-bind
//%arg0%.length
func Len(subject any) int {}

//js-bind
//setInterval(%args%)
func SetInterval(f func(), delay int) int {}

//js-bind
//clearInterval(%args%)
func ClearInterval(id int) {}
