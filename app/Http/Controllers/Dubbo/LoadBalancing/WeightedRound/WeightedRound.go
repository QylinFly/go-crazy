package WeightedRound
     
import (
	"sync"
	// "strconv"
)
 

type WeightedRound struct{
	index int  //表示上一次选择的服务器
	cw int  //表示当前调度的权值
	gcd int  //当前所有权重的最大公约数 比如 2，4，8 的最大公约数为：2
	mutex sync.Mutex
	slaveHosts map[int]map[string]interface{}
}

func NewWithWeights(weights map[string]int) *WeightedRound {

	weightedRound := &WeightedRound{
	}
	weightedRound.UpdateWeights(weights)
	return weightedRound
}

func (h *WeightedRound) UpdateWeights(weights map[string]int) *WeightedRound {

	h.mutex.Lock()
	var idx int = 0
	h.slaveHosts = make(map[int]map[string]interface{})
	numList :=  make(map[int]int)
	for key, w := range weights {
		h.slaveHosts[idx] = make(map[string]interface{})
		h.slaveHosts[idx]["host"] = key
		h.slaveHosts[idx]["weight"] = w
		numList[idx] = w
		idx++
	}

	if len(weights) > 0 {
		h.gcd = GcdArray(numList)
		h.index = 0
		h.cw = h.slaveHosts[0]["weight"].(int)
	}

	h.mutex.Unlock()
	return h
}

func (h *WeightedRound) GetNode(stringKey string) (node string, ok bool) {
	stringKey = ""

	if len(h.slaveHosts) == 0{
		return "" , false
	}

	defer h.mutex.Unlock()
	h.mutex.Lock()

	for{
		h.index = (h.index + 1) % len(h.slaveHosts)
		if h.index == 0 {
			h.cw = h.cw - h.gcd
			if h.cw <= 0 {
				h.cw = h.GetMaxWeight()
				if h.cw == 0 {
					return "" , false
				}
			}
		}

		if weight, _ := h.slaveHosts[h.index]["weight"].(int); weight >= h.cw {
			return h.slaveHosts[h.index]["host"].(string) , true
		}
	}
	return "" , false
}

func (h *WeightedRound) GetMaxWeight() int {
	max := 0
	for _, v := range h.slaveHosts {
		if weight, _ := v["weight"].(int); weight >= max {
			max = weight
		}
	}
	return max
}

// var slaveHosts = map[int]map[string]interface{}{
// 	0: {"host": "127.0.0.1", "weight": 2},
// 	1: {"host": "127.0.0.1", "weight": 4},
// 	2: {"host": "127.0.0.1", "weight": 8},
// }
 



// func getMaxWeight() int {
// 	max := 0
// 	for _, v := range slaveHosts {
// 		if weight, _ := v["weight"].(int); weight >= max {
// 			max = weight
// 		}
// 	}
// 	return max
// }


//Func to implement Euclid Algo
func Gcd(x, y int) int {
	for y != 0 {
	 x, y = y, x%y
	}
	return x
}

func GcdArray( numList map[int]int) int{
	var n int = len(numList)
	var result int

	//This is the result for only 2 integers
	result = Gcd(numList[1], numList[2])

	//for loop in case there're more than 2 ints
	for j := 3; j <= n; j++ {
	result = Gcd(result, numList[j])
	}

	return result
}

 
// func test() {

// 	note := map[string]int{}
 
// 	for i := 0; i < 100; i++ {
// 		s := getDns()
// 		fmt.Println(s)
// 		if note[s] != 0 {
// 			note[s]++
// 		} else {
// 			note[s] = 1
// 		}
// 	}
 
// 	for k, v := range note {
// 		fmt.Println(k, " ", v)
// 	}
// }
