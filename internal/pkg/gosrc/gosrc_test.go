package gosrc

import "testing"

func TestNewGoSrcFromModFile(t *testing.T) {
	gos, err := NewGoSrcFromModFile("../../../go.mod")
	if err != nil {
		t.Fatal(err)
	}

	path, exist, err := gos.GetAbsPath("github.com/zbysir/gopenapi/internal/model")
	if err != nil {
		t.Fatal(err)
	}
	// Z:\golang\go_project\gopenapi\internal\model
	t.Logf("%+v %+v", exist, path)


	path, exist, err = gos.GetAbsPath("github.com/zbysir/gopenapi/internal/delivery/http/handler/pet.go")
	if err != nil {
		t.Fatal(err)
	}
	// Z:\golang\go_project\gopenapi\internal\delivery\http\handler\pet.go
	t.Logf("%+v %+v", exist, path)
}
