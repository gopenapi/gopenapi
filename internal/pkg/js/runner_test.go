package js

import (
	"github.com/zbysir/gopenapi/internal/pkg/goast"
	"go/ast"
	"gopkg.in/yaml.v2"
	"log"
	"testing"
)

func TestRunJs(t *testing.T) {
	v, err := RunJs("[...model.FindPetByStatusParams, {name: 'status', required: true}]", func(name string, want string) interface{} {

		gp:=goast.NewGoParse()

		goStruct,exist,err:=gp.GetType("../../../"+name)
		if err != nil {
			t.Errorf("%v", err)
		}


		switch want {
		case "slice":
			return []interface{}{
				map[string]interface{}{
					"name": "id",
				},
			}
		}
		log.Printf("getter name: %+v %v", name, want)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}


	t.Logf("%T %v", v, v)
}

//
func goType2Params(t *goast.Type) ([]interface{}, error) {
	switch t.Type {
	case "struct":
		var r []interface{}
		for _, f := range t.Fields {
			r = append(r, map[string]interface{}{
				"name":
			})
		}
	}
}

func goType2Schema(t *goast.Type) (map[string]interface{}, error){
	 r:= map[string]interface{}{}
	switch t:=t.Type.(type) {
	case *ast.StructType:
		r["type"] = "object"
		prop:=[]yaml.MapItem
		for t.Fields
		r["properties"] =
	}
}