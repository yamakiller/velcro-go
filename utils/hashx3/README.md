# XXH3 hash algorithm
64/128位xxh3算法的Go实现，添加了SIMD向量指令集: AVX2 和 SSE2 支持加速哈希处理。\
原始库可以在这里找到: https://github.com/Cyan4973/xxHash


## 概述

对于输入长度大于240的情况，64位版本的xxh3算法按照以下步骤得到哈希结果。\
如果输入数据大小不大于240字节，计算步骤与上述类似。主要区别在于数据对齐。在输入较小的情况下，对齐大小为 16 字。


## 快速使用
SIMD汇编文件可以通过以下命令生成:
```
cd internal/avo && ./build.sh
```

Use Hash functions in your code:
```
package main

import "github.com/yamakiller/velcro-go/utils/hashx3"

func main() {
    println(hashx3.HashString("hello world"))
    println(hashx3.Hash128String("hello world!"))
}
```