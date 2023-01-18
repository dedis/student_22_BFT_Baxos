package polypus

import (
	"fmt"
	"go.dedis.ch/dela/core"
	"student_22_BFT_Baxos/polypus/lib/polylib"
)

func InitPoly(basePolyPort int) (*core.Watcher, *core.Watcher) {
	//polypus
	inWatch := core.NewWatcher()
	outWatch := core.NewWatcher()

	polyAddr := fmt.Sprintf("localhost:%d", basePolyPort)
	proxy := polylib.Start(polyAddr, inWatch, outWatch, nil)
	polyConfig := fmt.Sprintf(`{"id": "%s", "addr": "%s", "proxy": "http://%s"}`, "AC", "client", proxy)
	fmt.Println(polyConfig)

	return inWatch, outWatch
}
