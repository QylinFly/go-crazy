/*
 * 
 * File: loadbalancingctrl.go
 * Author: QylinFly (18612116114@163.com)
 * Created: 星期 5, 2018-5-18 11:34:55 am
 * -----
 * Modified By: QylinFly (18612116114@163.com>)
 * Modified: 星期 5, 2018-5-18 11:35:19 am
 * -----
 * Copyright 2017 - 2027
 *	数值计算参考
 *	https://www.gonum.org/post/intro-to-stats-with-gonum/
 *	一致性哈希
 *	https://github.com/serialx/balancedStrategy
 */

 package LoadBalancing
 
 import (
	"fmt"
	"math"
	"sort"
	"time"
	"strconv"
	// "math/rand"
	// "github.com/satori/go.uuid"
	"gonum.org/v1/gonum/stat"
	"github.com/xoxo/crm-x/util/logger"
	// "github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/LoadBalancing/hashring"
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/LoadBalancing/WeightedRound"	
	"github.com/xoxo/crm-x/app/Http/Controllers/Dubbo/Util"
)

type LoadBalancingCtrl struct {
	weights    map[string]int
	endpoints  map[string]int
	// 平衡策略
	balancedStrategy 	*WeightedRound.WeightedRound//*hashring.HashRing
	responseInfo map[string][]int
	recordEvents chan RecordStruct
}

func New(nodes []*Util.Endpoint) *LoadBalancingCtrl {
	weights := make(map[string]int)
	endpoints := make(map[string]int)
	for _, node := range nodes {
		weights[node.ToString()] = node.Weight()
		endpoints[node.ToString()] = node.Weight()
	}
	loadBalancing := newWithWeights(endpoints,endpoints)
	return loadBalancing
}

func newWithWeights(endpoints map[string]int,weights map[string]int) *LoadBalancingCtrl {
	loadBalancing := &LoadBalancingCtrl{
		weights		: weights,
		endpoints	: endpoints,
		responseInfo: make(map[string][]int),
		recordEvents: make(chan RecordStruct,1024),
	}
	loadBalancing.balancedStrategy = WeightedRound.NewWithWeights(loadBalancing.weights) //balancedStrategy.NewWithWeights(loadBalancing.weights)
	return loadBalancing
}


// func randGenerator()  string {
//     return uuid.NewV4().String()
// }

func (ctrl * LoadBalancingCtrl) Get() string{
	// uuid := randGenerator()
	// logger.Info("uuid = "+uuid)
	// uuid2 := strconv.Itoa( time.Now().Nanosecond()+rand.Int() )
	node,_ := ctrl.balancedStrategy.GetNode("uuid2")
	return node
}

func (ctrl * LoadBalancingCtrl) GetByKey(stringKey string) string{
	node,_ := ctrl.balancedStrategy.GetNode(stringKey)
	return node
}

func (ctrl * LoadBalancingCtrl) Update(nodes []*Util.Endpoint) *LoadBalancingCtrl{
	weights := make(map[string]int)
	for _, node := range nodes {
		weights[node.ToString()] = node.Weight()
	}
	ctrl.balancedStrategy.UpdateWeights(weights)
	return ctrl
}

type RecordStruct struct {
	node string
    time int
}

func (ctrl *LoadBalancingCtrl)RecordResponseInfo(node string,latency int ){
	r := RecordStruct{node,latency}
	ctrl.recordEvents <- r
}
func (ctrl *LoadBalancingCtrl)recordResponseInfo(node string,latency int ){
	ctrl.responseInfo[node] = append(ctrl.responseInfo[node], latency)
}

func (ctrl *LoadBalancingCtrl)PrintRecordResponseInfo(){
	for key, node := range ctrl.responseInfo {
		logger.Info("调用情况："+key+" count="+ strconv.Itoa(len(node)))
	}
}

//  响应统计
func (agent *LoadBalancingCtrl) RecordResponseInfoLoop(){
	for { // Check for updates
		select {
		case event := <- agent.recordEvents:
			agent.recordResponseInfo(event.node,event.time)
		case <-time.After(10 * time.Second):
			agent.PrintRecordResponseInfo()
		}
	}
}

func test() {
	xs := []float64{
		32.32, 56.98, 21.52, 44.32,
		55.63, 13.75, 43.47, 43.34,
		12.34,
	}

	fmt.Printf("data: %v\n", xs)

	// computes the weighted mean of the dataset.
	// we don't have any weights (ie: all weights are 1)
	// so we just pass a nil slice.
	mean := stat.Mean(xs, nil)
	variance := stat.Variance(xs, nil)
	stddev := math.Sqrt(variance)

	// stat.Quantile needs the input slice to be sorted.
	sort.Float64s(xs)
	fmt.Printf("data: %v (sorted)\n", xs)

	// computes the median of the dataset.
	// here as well, we pass a nil slice as weights.
	median := stat.Quantile(0.5, stat.Empirical, xs, nil)

	fmt.Printf("mean=     %v\n", mean)
	fmt.Printf("median=   %v\n", median)
	fmt.Printf("variance= %v\n", variance)
	fmt.Printf("std-dev=  %v\n", stddev)
}