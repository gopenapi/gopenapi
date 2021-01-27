package js

import (
	rice "github.com/GeertJohan/go.rice"
	"github.com/dop251/goja"
	"github.com/mitchellh/mapstructure"
	"sync"
)

//go:generate rice embed-go

// 将es6语法转为es5

type babel struct {
	vm        *goja.Runtime
	this      goja.Value
	transform goja.Callable
	mutex     sync.Mutex // TODO: cache goja.CompileAST() in an init() function?
}

var once sync.Once
var globalBabel *babel

func newBabel() (*babel, error) {
	var err error

	once.Do(func() {
		conf := rice.Config{
			LocateOrder: []rice.LocateMethod{rice.LocateEmbedded},
		}
		babelSrc := conf.MustFindBox("lib").MustString("babel.min.js")
		vm := goja.New()
		if _, err = vm.RunString(babelSrc); err != nil {
			return
		}

		this := vm.Get("Babel")
		bObj := this.ToObject(vm)
		globalBabel = &babel{vm: vm, this: this}
		if err = vm.ExportTo(bObj.Get("transform"), &globalBabel.transform); err != nil {
			return
		}
	})

	return globalBabel, err
}

var DefaultOpts = map[string]interface{}{
	"presets":       []interface{}{"es2015"},
	"ast":           false,
	"sourceMaps":    false,
	"babelrc":       false,
	"compact":       false,
	"retainLines":   true,
	"highlightCode": false,
}

func (b *babel) Transform(src, filename string) (string, *SourceMap, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	opts := make(map[string]interface{})
	for k, v := range DefaultOpts {
		opts[k] = v
	}
	opts["filename"] = filename

	v, err := b.transform(b.this, b.vm.ToValue(src), b.vm.ToValue(opts))
	if err != nil {
		return "", nil, err
	}

	vO := v.ToObject(b.vm)
	var code string
	if err = b.vm.ExportTo(vO.Get("code"), &code); err != nil {
		return code, nil, err
	}
	var rawMap map[string]interface{}
	if err = b.vm.ExportTo(vO.Get("map"), &rawMap); err != nil {
		return code, nil, err
	}
	var srcMap SourceMap
	if err = mapstructure.Decode(rawMap, &srcMap); err != nil {
		return code, &srcMap, err
	}
	return code, &srcMap, err
}

type SourceMap struct {
	Version    int
	File       string
	SourceRoot string
	Sources    []string
	Names      []string
	Mappings   string
}

func Transform(src, filename string) (code string, srcmap *SourceMap, err error) {
	var b *babel
	if b, err = newBabel(); err != nil {
		return
	}

	return b.Transform(src, filename)
}
