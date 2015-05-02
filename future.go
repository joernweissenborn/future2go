package future2go

import "sync"

func New() (F *Future) {
	F = new(Future)
	F.m = new(sync.Mutex)
	F.fcs = []futurecompleter{}
	F.fces = []futurecompletererror{}
	return
}


type CompletionFunc func(interface {}) interface {}
type ErrFunc func(error) (interface {}, error)

type Future struct {
	m *sync.Mutex
	fcs []futurecompleter
	fces []futurecompletererror
	c bool
	r interface {}
	e error
}

func (f *Future) Complete(Data interface {}){
	f.m.Lock()
	defer f.m.Unlock()
	f.r = Data
	for _, fc := range f.fcs {
		deliverData(fc,Data)
	}
	f.c = true
}

func (f *Future) IsComplete() bool {
	return f.c
}

func (f *Future) CompleteError(err error){
	f.m.Lock()
	defer f.m.Unlock()
	f.e = err
	for _, fce := range f.fces {
		deliverErr(fce,f.e)
	}
	f.c = true
}

func deliverData(fc futurecompleter, d interface {}){
	go func(){
		fc.f.Complete(fc.cf(d))
	}()
}

func deliverErr(fce futurecompletererror, e error){
	go func(){
		d, err := fce.ef(e)
		if err == nil {
			fce.f.Complete(d)
		} else {
			fce.f.CompleteError(err)
		}
	}()
}

func (f *Future) Then(cf CompletionFunc) (nf *Future) {
	f.m.Lock()
	defer f.m.Unlock()

	nf = New()
	fc := futurecompleter{cf,nf}
	if f.c && f.e == nil {
		deliverData(fc,f.r)
	} else if !f.c  {
		f.fcs = append(f.fcs,fc)
	}
	return
}

func (f *Future) WaitUntilComplete() {
	c := make(chan struct{})
	defer close(c)
	cmpl := func(interface {})interface {}{
		c<-struct{}{}
		return nil
	}
	f.Then(cmpl)
	f.Err(cmpl)
	<-c
}

func (f *Future) GetResult() interface {}{
	return f.r
}

func (f *Future) Err(ef ErrFunc) (nf *Future) {
	f.m.Lock()
	defer f.m.Unlock()

	nf = New()
	fce := futurecompletererror{ef, nf}
	if f.e != nil {
		deliverErr(fce, f.e)
	} else if !f.c{
		f.fces = append(f.fces, fce)
	}
	return
}

type futurecompleter struct {
	cf CompletionFunc
	f *Future
}

type futurecompletererror struct {
	ef ErrFunc
	f *Future
}
