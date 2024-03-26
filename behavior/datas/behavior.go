package datas

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Behavior3Node behavior3的节点
type Behavior3Node struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Children    []string               `json:"children"`
	Child       string                 `json:"child"`
	Properties  map[string]interface{} `json:"properties"`
}

// Behavior3Tree behavior3树
type Behavior3Tree struct {
	ID          string                    `json:"id"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	Root        string                    `json:"root"`
	Properties  map[string]interface{}    `json:"properties"`
	Nodes       map[string]*Behavior3Node `json:"nodes"`
}

// Behavior3Project behavior3的工程json类型
type Behavior3Project struct {
	SelectedTree string           `json:"selectedTree"`
	Scope        string           `json:"scope"`
	Trees        []*Behavior3Tree `json:"trees"`
}

// Behavior3File 是behavior3编辑器保存的.b3格式的配置文件
type Behavior3File struct {
	Name string           `json:"name"`
	Data Behavior3Project `json:"data"`
}

func LoadB3File(path string) (*Behavior3File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseB3File(data)
}

// ParseB3File parse project from []byte
func ParseB3File(data []byte) (*Behavior3File, error) {
	var proj Behavior3File
	err := json.Unmarshal(data, &proj)
	if err != nil {
		return nil, err
	}
	return &proj, nil
}

func (node *Behavior3Node) GetFloat64(name string) float64 {
	v := node.Properties[name]
	if v == nil {
		return 0
	}
	f,err := strconv.ParseFloat(v.(string),64)
	if err != nil{
		return 0
	}
	return f
}

func (node *Behavior3Node) GetInt(name string) int {
	return int(node.GetFloat64(name))
}

func (node *Behavior3Node) GetInt32(name string) int32 {
	return int32(node.GetFloat64(name))
}

func (node *Behavior3Node) GetInt64(name string) int64 {
	return int64(node.GetFloat64(name))
}

func (node *Behavior3Node) GetUint32(name string) uint32 {
	return uint32(node.GetFloat64(name))
}

func (node *Behavior3Node) GetUint64(name string) uint64 {
	return uint64(node.GetFloat64(name))
}

func (node *Behavior3Node) GetString(name string) string {
	return (node.Properties[name]).(string)
}

func (node *Behavior3Node) GetBool(name string) bool {
	return (node.Properties[name]).(bool)
}

func (node *Behavior3Node) GetInt32s(name string) []int32 {
	var ret []int32
	v := node.Properties[name]
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		if v == "" {
			return nil
		}
		lst := strings.Split(val, ",")
		for _, val := range lst {
			n, err := strconv.Atoi(val)
			if err != nil {
				panic(fmt.Errorf("failed to unmarshal int32s: %v", err))
			}
			ret = append(ret, int32(n))
		}
	case float64:
		ret = append(ret, int32(val))
	}

	return ret
}

func (node *Behavior3Node) GetInt64s(name string) []int64 {
	var ret []int64
	v := node.Properties[name]
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		if v == "" {
			return nil
		}
		lst := strings.Split(val, ",")
		for _, val := range lst {
			n, err := strconv.Atoi(val)
			if err != nil {
				panic(fmt.Errorf("failed to unmarshal int64s: %v", err))
			}
			ret = append(ret, int64(n))
		}
	case float64:
		ret = append(ret, int64(val))
	}

	return ret
}
