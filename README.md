# delayEncode

## Parameters

- T: Given time, the time to generate data must be greater than this given time;

- L: Length of generated data;

- R: Speed ​​of generated data;

- B: Length of data block;

- f(*): Difficulty function;

- h(*): Hash function;

- g(*): Function to generate dependency

- s: Seed to generate data;

## Encoding

The purpose of each of the first m layers is to increase the difficulty, and it cannot be calculated in reverse or in parallel. The purpose of using multiple layers is to prevent pre-storing part of the data, so that the encoding time is less than $T$.

### layer 1

The purpose of layer 0 is to increase the difficulty, and it cannot be calculated in reverse or in parallel.
$$
k_0^1 = f(s) \\k_1^1 = f(\{k_l;l\in g(1,c)\}) \\k_2^1 = f(\{k_l;l\in g(2,c)\}) \\... \\k_j^1 = f(\{k_l;l\in g(j,c)\}) \\... \\k_{n-1}^1 = f(\{k_l;l\in g(n-1,c)\}) \\k_{n}^1 = f(\{k_l;l\in g(n,c)\}) \\
$$

### layer i

Odd layers are from 0 to n-1, and even layers are from n-1 to 0; the generated dependency order, the initial value is the result of the same position in the previous layer, and depends on the data already calculated in this layer.
$$
k_0^1 = f(s) \\k_1^1 = f(k_1^{l-1},\{k_l;l\in g(1,c)\}) \\k_2^1 = f(k_2^{l-1},\{k_l;l\in g(2,c)\}) \\... \\k_j^1 = f(k_j^{l-1},\{k_l;l\in g(j,c)\}) \\... \\k_{n-1}^1 = f(k_{n-1}^{l-1},\{k_l;l\in g(n-1,c)\}) \\k_{n}^1 = f(k_n^{l-1},\{k_l;l\in g(n,c)\}) \\
$$

### layer m

The purpose of the last layer is to distribute the difficulty to each bit of data.
$$
e_n = k_n \\e_{n-1} = (e_n + k_{n-1}) \bigoplus k_0 \\... \\e_j = (e_{j+1} + k_j) \bigoplus k_{n-1-j} \\...\\e_1 = (e_2 + k_1) \bigoplus k_{n-2} \\e_0 = (e_1+k_0) \bigoplus k_{n-1}
$$

### Function selection

- Difficulty function $f$, for example $f(x) = x^{2^t}$, t is the difficulty value, that is, the larger $t$ is, the slower $R$ is.
- Hash function h, choose hash256;
- Dependent function g, choose to use [DRSample](https://acmccs.github.io/papers/p1001-alwenA.pdf), and filecoin's modified algorithm [DRG](https://github.com/filecoin-project/drg-attacks/blob/master/notes.md)

The following is pseudo code:

```
// i is the i-th data slice (32B or 128B, etc.)
// c means generating c parents, that is, the current $k_i$ calculation depends on the results of c previous $k_{?}$
// s is the initial id of block production; k_{x} is the x-th calculation result
g(i,c) {
r = rand.NewRander(s+i) //Generate a new random seed
logi = math.Floor(math.log2(i*c))

if i == 0 {
return 0
}

// Calculate each dependency
for l:=0; l<c;l++ {
j = r.Next() % logi
jj = math.Min(i*c+l, 1 << (j + 1))
dist = r.Next(math.Max(jj >> 1, 2), jj + 1);
if i - dist == i {
dag[l] = i-1
} else {
dag[l] = i - dist // dag saves each dependency
}
}

return dag
}

// x is the input value
h(x) {
return hash256(x)
}

// x is the input value
// t is the difficulty function
// l represents the lth layer
f([]x) {
// Sum the input
res = sum(x)
hx = hash(res)
t = exp(2,t)
// You can choose to perform exponential operation on the hash result or on x.
return add(res,exp(hx,t))
}
```

## Data generation analysis

Data is generated incrementally according to the seed, and d bits of data are generated each time. Each d-bit generation includes:

(c+1)*(m-1)+1 ADD operation; 1 XOR operation; m-1 EXP operation; m-1 hash operation; m-1 g function call;

Test results, running on i5-8250 @1.6G cpu

| d | ADD | XOR | EXP | hash | g | encode(c=2) |
| :--------: | :---: | :---: | :-------------------: | :---: | :---: | :---------: |
| 4096(m=2) | 200ns | 200ns | 50us(t = 4, base is 512B) | 450ns | 400ns | 7.4 MB/s |
| 4096(m=2) | 200ns | 200ns | 5us(t = 4, base is 32B) | 450ns | 400ns | 63 MB/s |
| 4096(m=11) | 200ns | 200ns | 5us(t = 4, base is 32B) | 450ns | 400ns | 8.1 MB/s |

## Forgery Difficulty Analysis

Assuming that $\delta\times L(\delta < 1)$ data is pre-stored, when using (m+1) layer encoding, the data generation time is $T_e$ and the calculation time is $T_e\times (1-\delta/m)$, when $\delta->1$, the minimum time is $T_e\times (1-1/m)$, that is, even if 1 bit of data is stored less, $T_e\times (1-1/m)$ is required; in this case, $T_e\times (1-1/m) > T$.

## Performance Analysis

$R$ is mainly limited by the difficulty function, and $R$ needs to meet the following conditions:

$T \lt \frac{L}{R}\times (1-\frac{1}{m})$

Therefore, when $T$ is kept unchanged, the smaller the difficulty, the larger $R$, and the correspondingly larger the generated data size.

- Write performance: Divide the user space into fixed size D, and each block has a seed for generating data. The size of the generated data is D, and the write performance is $R$;
- Read performance: Users read continuously, and the maximum performance is $R$;
- Repair performance: The time to repair a block is $T$, so the larger the block, the better the repair performance;
