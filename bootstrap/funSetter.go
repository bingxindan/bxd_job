package bootstrap

type FuncSetter struct {
	beforeFuncs []BeforeServerStartFunc
	afterFuncs []AfterServerStopFunc
}

func NewFuncSetter() *FuncSetter {
	return &FuncSetter{}
}

func (f *FuncSetter) AddBeforeServerStartFunc(fns ...BeforeServerStartFunc) {
	for _, fn := range fns {
		f.beforeFuncs = append(f.beforeFuncs, fn)
	}
}

func (f *FuncSetter) AddAfterServerStopFunc(fns ...AfterServerStopFunc) {
	for _, fn := range fns {
		f.afterFuncs = append(f.afterFuncs, fn)
	}
}

func (f *FuncSetter) RunBeforeServerStartFunc() error {
	for _, fn := range f.beforeFuncs {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FuncSetter) RunAfterServerStopFunc() {
	for _, fn := range f.afterFuncs {
		fn()
	}
}
