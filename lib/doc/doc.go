package doc

type HTMLElement struct {}

//js-bind
//document.querySelector(%args%)
func QuerySelector(query string) HTMLElement {}

//js-bind
//document.querySelectorAll(%args%)
func QuerySelectorAll(query string) []HTMLElement {}

//js-bind
//document.createElement(%args%)
func CreateElement(name string) HTMLElement {}

//js-bind
//%recv%[%arg0%] = %arg1%;
func (HTMLElement) Set(field string, value string) {}

//js-bind
//%recv%.setAttribute(%args%)
func (HTMLElement) SetAttr(attr string, value string) {}

//js-bind
//%recv%.appendChild(%args%)
func (HTMLElement) AppendChild(elem HTMLElement) {}

//js-bind
//%recv%.play(%args%)
func (HTMLElement) Play() {}
