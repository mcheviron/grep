package worklist

type Worklist chan string

func NewWorklist(capacity int) Worklist {
	return make(Worklist, capacity)
}
func (w *Worklist) Add(path string) {
	*w <- path
}
func (w *Worklist) Get() string {
	return <-*w
}

type Result struct {
	Line    string
	LineNum int
	Path    string
}

type Results chan Result

func NewResults(capacity int) Results {
	return make(Results, capacity)
}

func (r *Results) Add(line string, lineNum int, path string) {
	*r <- Result{line, lineNum, path}
}
func (r *Results) Get() Result {
	return <-*r
}
