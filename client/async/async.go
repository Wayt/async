package async

var DefaultEngine = NewEngine()

func Func(name string, fun Function) {
	DefaultEngine.Func(name, fun)
}

func Run() error {
	return DefaultEngine.Run()
}
