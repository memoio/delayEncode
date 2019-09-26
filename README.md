# delayEncode

## 参数

- T：给定时间，生成数据的时间要大于此给定时间；
- L：生成的数据的长度；
- R：生成数据的速度；
- B：数据块的长度；
- f(*)：难度函数；
- h(*)：哈希函数；
- g(*)：生成依赖关系的函数
- s：生成数据的种子；

## 编码

前面m层每一层的目的是增加难度，不能逆向计算而且不能并行计算。采用多层目的是防止预存一部分数据，使得encoding时间小于$T$。

### layer  1

第0层目的是增加难度，不能逆向计算而且不能并行计算。
$$
k_0^1 = f(s) \\k_1^1 = f(\{k_l;l\in g(1,c)\}) \\k_2^1 = f(\{k_l;l\in g(2,c)\}) \\... \\k_j^1 = f(\{k_l;l\in g(j,c)\}) \\... \\k_{n-1}^1 = f(\{k_l;l\in g(n-1,c)\})  \\k_{n}^1 = f(\{k_l;l\in g(n,c)\}) \\
$$

### layer i

奇数层从0到n-1，偶数层从n-1到0；生成的依赖顺序，初始值为上一层同一位置的结果，依赖本层已经算出来的数据。
$$
k_0^1 = f(s) \\k_1^1 = f(k_1^{l-1},\{k_l;l\in g(1,c)\}) \\k_2^1 = f(k_2^{l-1},\{k_l;l\in g(2,c)\}) \\... \\k_j^1 = f(k_j^{l-1},\{k_l;l\in g(j,c)\}) \\... \\k_{n-1}^1 = f(k_{n-1}^{l-1},\{k_l;l\in g(n-1,c)\})  \\k_{n}^1 = f(k_n^{l-1},\{k_l;l\in g(n,c)\}) \\
$$


### layer  m

最后一层的目的是将难度分散到每一位数据上。
$$
e_n = k_n \\e_{n-1} = (e_n + k_{n-1}) \bigoplus k_0 \\... \\e_j = (e_{j+1} + k_j) \bigoplus k_{n-1-j} \\...\\e_1 = (e_2 + k_1) \bigoplus k_{n-2} \\e_0 = (e_1+k_0) \bigoplus k_{n-1}
$$

### 函数选择



- 难度函数$f$，例如$f(x) = x^{2^t}$，t为 难度值即$t$越大，$R$越慢。
- 哈希函数h，选择hash256;
- 依赖函数g，选择使用[DRSample](https://acmccs.github.io/papers/p1001-alwenA.pdf)，以及filecoin修改后的算法[DRG](https://github.com/filecoin-project/drg-attacks/blob/master/notes.md)

以下为伪代码：

```
// i为第i个数据片（32B或128B等）
// c表示产生c个parents，即当前的$k_i$计算依赖c个前面的$k_{?}$的结果
// s为块生产的初始id;k_{x}为第x个计算结果
g(i,c) {
	r = rand.NewRander(s+i) //产生新的随机种子
	logi = math.Floor(math.log2(i*c))
	
	if i == 0 {
		return 0
	}
	
	// 计算每一个依赖
	for l:=0; l<c;l++ {
		j = r.Next() % logi
		jj = math.Min(i*c+l, 1 << (j + 1))
		dist = r.Next(math.Max(jj >> 1, 2), jj + 1);
		if i - dist == i {
			dag[l] = i-1
		} else {
			dag[l] = i - dist // dag保存每一个依赖
		}
	}

	return dag
}

// x为输入值
h(x) {
	return hash256(x)
}

// x为输入值
// t为难度函数
// l表示第l层
f([]x) {
	// 对输入求和
	res = sum(x)
	hx = hash(res)
	t = exp(2,t)
	// 可以选择对hash结果算指数运算，也可以对x算hash结果
	return add(res,exp(hx,t)) 
}
```



## 数据生成分析

数据根据种子递增生成，每次生成d位数据，每次d位的生成包括：

（c+1）*(m-1)+1次ADD操作；1次XOR操作；m-1次EXP操作；m-1次hash操作；m-1次g函数调用;

测试结果，在i5-8250 @1.6G cpu上运行

|     d      |  ADD  |  XOR  |          EXP          | hash  |   g   | encode(c=2) |
| :--------: | :---: | :---: | :-------------------: | :---: | :---: | :---------: |
| 4096(m=2)  | 200ns | 200ns | 50us(t = 4, 底为512B) | 450ns | 400ns |  7.4 MB/s   |
| 4096(m=2)  | 200ns | 200ns |  5us(t = 4, 底为32B)  | 450ns | 400ns |   63 MB/s   |
| 4096(m=11) | 200ns | 200ns |  5us(t = 4, 底为32B)  | 450ns | 400ns |  8.1 MB/s   |



## 伪造难度分析

假设预存$\delta\times L(\delta < 1)$的数据，在使用(m+1) 层编码的时候，生成数据时间为$T_e$，计算时间为$T_e\times (1-\delta/m)$，在$\delta->1$的时候，最小时间为$T_e\times (1-1/m)$，也就是说即使少存1bit的数据，也需要$T_e\times (1-1/m)$；此时需要$T_e\times (1-1/m) > T$即可。

## 性能分析

$R$主要被难度函数限制，$R$需要满足条件：

$T \lt \frac{L}{R}\times (1-\frac{1}{m})$

因此维持$T$不变的时候，难度选择越小，$R$越大，相应的生成的数据大小也越大。

- 写入性能：将用户空间按固定大小D划分，每一块一个生成数据的种子，生成的数据大小为D，则写入性能为$R$；
- 读取性能：用户连续读取，最大性能为$R$；
- 修复性能：修复一个块的时间$T$，因此块越大，修复性能越好；

