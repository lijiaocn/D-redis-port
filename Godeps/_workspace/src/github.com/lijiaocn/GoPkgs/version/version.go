//create: 2015/05/14 15:55:20 change: 2015/09/16 15:08:57 author:lijiao
package version

import(
	"fmt"
)

var(
	VERSION string
	COMPILE string
)

func Show(){
	fmt.Printf("version: %s   compile at: %s\n", VERSION, COMPILE)
}
