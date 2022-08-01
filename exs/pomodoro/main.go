package main

import (
    . "lib/std"
    "lib/fmt"
    "lib/doc"
    "lib/math"
)

const (
    ModeSession int = 0
    ModeBreak = 1
    ModeLongBreak = 2
)

type Person struct {
    name string
    age int
}

type Timer struct {
    time int
    mode int
    elem doc.HTMLElement
    running bool
    interval int
    sessions int
    whenOver func()
    // default time values
    defaultSession int
    defaultBreak int
    defaultLongBreak int
}

func CreateElement(a int) {
    fmt.Println(a)
}

func trailingZeros(s string, max int) string {
    if Len(s) < max {
        out := ""
        for i := 0; i < max - Len(s); i++ {
            out += "0"
        }
        return  out + s
    }
    return s
}

func (timer Timer) String() string {
    time := timer.time
    mins := math.FakeFloor(time / 60);
    secs := time - (mins * 60);
    return trailingZeros(Stringify(mins), 2) + ":" +
           trailingZeros(Stringify(secs), 2)
}

func (timer Timer) Init() {
    timer.Reset()
}

func (timer Timer) Reset() {
    timer.Pause()
    timer.SetTime()
    timer.elem.Set("innerText", timer.String())
}

func (timer Timer) Start() {
    if !timer.running {
        timer.running = true
        timer.interval = SetInterval(func() {
            timer.Update()
        }, 1000)
    }
}

func (timer Timer) Pause() {
    timer.running = false
    ClearInterval(timer.interval)
}

func (timer Timer) WhenOver(f func ()) {
    timer.whenOver = f
}

func (timer Timer) SetTime() {
    switch (timer.mode) {
        case ModeSession:
            timer.time = timer.defaultSession
            break
        case ModeBreak:
            timer.time = timer.defaultBreak
            break
        case ModeLongBreak:
            timer.time = timer.defaultLongBreak
            break
        default: {}
    }
}

func (timer Timer) SetMode(mode int) {
    timer.mode = mode
}

func (timer Timer) GetModeString() string {
    switch (timer.mode) {
        case ModeSession: return "session"
        case ModeBreak: return "break"
        case ModeLongBreak: return "long break"
        default: return "(unknown)"
    }
}

func (timer Timer) Update() {
    if timer.time > 0 {
        timer.time -= 1
    } else {
        if timer.mode == ModeSession {
            timer.sessions++
            if timer.sessions == 3 {
                timer.sessions = 0
                timer.SetMode(ModeLongBreak)
            } else {
                timer.SetMode(ModeBreak)
            }
        } else {
            timer.SetMode(ModeSession)
        }
        timer.Reset()
        timer.whenOver()
    }
    timer.elem.Set("innerText", timer.String())
}

func Button(text string, onclick string) doc.HTMLElement {
    e := doc.CreateElement("button")
    e.Set("innerText", text)
    e.SetAttr("class", "timer__play_button")
    e.SetAttr("onclick", onclick)
    return e
}

func timerContainer() doc.HTMLElement {
    e := doc.CreateElement("div")
    e.SetAttr("class", "timer__container")
    return e
}

func root() doc.HTMLElement {
    e := doc.QuerySelector("#root")
    return e
}

func main() {
    root := root()
    container := timerContainer()
    title := doc.CreateElement("p")
    timer := Timer{
        time: 0,
        mode: ModeSession,
        elem: doc.CreateElement("p"),
        running: false,
        interval: 0,
        sessions: 0, 
        whenOver: func() {},
        defaultSession: 60 * 25,
        defaultBreak: 60 * 5,
        defaultLongBreak: 60 * 15,
    }

    buttons := doc.CreateElement("div")
    buttons.SetAttr("class", "timer__buttons")

    buttons.AppendChild(Button("start", "timer.Start()"))
    buttons.AppendChild(Button("pause", "timer.Pause()"))
    buttons.AppendChild(Button("reset", "timer.Reset()"))

    container.AppendChild(title)
    container.AppendChild(timer.elem)
    container.AppendChild(buttons)

    root.AppendChild(container)

    timer.Init()
    timer.WhenOver(func() {
        title.Set("innerText", timer.GetModeString())
    })

    title.Set("innerText", timer.GetModeString())
}
