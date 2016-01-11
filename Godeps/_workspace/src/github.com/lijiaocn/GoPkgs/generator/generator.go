//create: 2015/09/17 16:24:20 change: 2015/09/17 16:25:18 author:lijiao
package generator

func StrGenerator(strs []string) func()string{
	next := 0
	gen := func()string{
		if next == len(strs) - 1{
			str := strs[next]
			next = 0
			return str
		}
		str := strs[next]
		next++
		return str
	}
	return gen
}
