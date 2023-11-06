/*
@File    :   KDTree.go
@Time    :   2023/11/06 16:25:05
@Author  :   Skyrim
@Version :   1.0
@Site    :   https://github.com/skyrimforest
@Desc    :   None
*/

package KDTree

// package main

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

// KDTreeNode树节点类
type KDTreeNode struct {
	data  []float64
	left  *KDTreeNode
	right *KDTreeNode
}

// KDTree 树类
type KDTree struct {
	dims          int
	discrete_dims map[int]int
	root          *KDTreeNode
}

// KDTreeNode 初始化
func NewKDTreeNode(data []float64) (*KDTreeNode, error) {
	return &KDTreeNode{
		data:  data,
		left:  nil,
		right: nil,
	}, nil
}

// KDTree 初始化
func NewKDTree(dims int, discrete_dims map[int]int) (*KDTree, error) {
	if dims < 0 {
		return nil, errors.New("KD tree data dimension must be positive")
	}
	for i, dis_dim := range discrete_dims {
		if dis_dim < 0 || dis_dim > dims {
			str := fmt.Sprintf("discrete_dim[{%d}]={%d} out of dim range", i, dis_dim)
			return nil, errors.New(str)
		}
	}
	return &KDTree{
		dims:          dims,
		discrete_dims: discrete_dims,
		root:          nil,
	}, nil
}

// 测试dim是否符合输入数据
func (kt *KDTree) check_dim(data []float64) error {
	if len(data) != kt.dims {
		str := fmt.Sprintf("Data dimension does not equal to specified dimension {%d}", kt.dims)
		return errors.New(str)
	}
	return nil
}

// 插入数据,左枝小于右枝
func (kt *KDTree) insert(data []float64, node *KDTreeNode, cut_dim int) *KDTreeNode {
	if node == nil {
		var err error
		node, err = NewKDTreeNode(data)
		if err != nil {
			log.Fatal(err)
		}
	} else if reflect.DeepEqual(data, node.data) {

	} else if data[cut_dim] < node.data[cut_dim] {
		node.left = kt.insert(data, node.left, (cut_dim+1)%kt.dims)
	} else {
		node.right = kt.insert(data, node.right, (cut_dim+1)%kt.dims)
	}
	return node
}

// 外部可用的插入
func (kt *KDTree) Insert(data []float64) {
	kt.check_dim(data)
	kt.root = kt.insert(data, kt.root, 0)
}

// 外部可用的映射操作
func MapForEach(data []float64, fn func(it float64) float64) []float64 {
	newArray := []float64{}
	for _, it := range data {
		newArray = append(newArray, fn(it))
	}
	return newArray
}

// 外部可用的规约操作
func ReduceForEach(data []float64, fn func(it float64) float64) float64 {
	var sum float64 = 0.0
	for _, it := range data {
		sum += fn(it)
	}
	return sum
}

// 计算向量距离
func (kt *KDTree) distance(data1, data2 []float64) float64 {
	data3 := make([]float64, len(data1))
	copy(data3, data1)
	target := make([]float64, len(data3))
	for i := 0; i < len(data3); i++ {
		target[i] = data3[i] - data2[i]
	}
	var res float64 = 0
	target = MapForEach(target, func(it float64) float64 {
		return it * it
	})
	res = ReduceForEach(target, func(it float64) float64 {
		return it
	})
	return res
}

// 递归计算最近的节点
func (kt *KDTree) get_nearest(data []float64, node *KDTreeNode, cut_dim int, cur_max *float64, cur_result *KDTreeNode) {
	if node == nil {
		return
	}
	cur_distance := kt.distance(data, node.data)
	if cur_distance >= *cur_max {
		return
	}
	*cur_max = cur_distance
	cur_result = node
	if data[cut_dim] < node.data[cut_dim] {
		kt.get_nearest(data, node.left, (cut_dim+1)%kt.dims, cur_max, cur_result)
		kt.get_nearest(data, node.right, (cut_dim+1)%kt.dims, cur_max, cur_result)
	} else {
		kt.get_nearest(data, node.right, (cut_dim+1)%kt.dims, cur_max, cur_result)
		kt.get_nearest(data, node.left, (cut_dim+1)%kt.dims, cur_max, cur_result)
	}
}

// 外部可用的计算最小值 返回最小距离与最小距离对应的向量
func (kt *KDTree) Get_nearest(data []float64) (float64, []float64) {
	kt.check_dim(data)
	cur_max := 1e8
	cur_result := kt.root
	kt.get_nearest(data, kt.root, 0, &cur_max, cur_result)

	var ret_tar []float64
	if cur_result == nil {
		ret_tar = nil
	} else {
		ret_tar = cur_result.data
	}
	return cur_max, ret_tar
}
