package delayencode

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"path"
)

//GenParams 利用生成参数，返回一个阶N以及对应的phi(N)
func GenParams(bits int) (*big.Int, *big.Int, error) {
	t := uint(bits)
	order := big.NewInt(1).Lsh(big.NewInt(1), t)
	phi := big.NewInt(1).Lsh(big.NewInt(1), t-2)
	return order, phi, nil
}

//以三副本为例,encode和decode为同一个函数
func layerEncode(stripeID []byte, data []byte, order *big.Int, bitLen, dif, layers int) []byte {
	v := big.NewInt(1).Lsh(big.NewInt(1), uint(dif))
	bigOne := new(big.Int).SetInt64(1)
	//需要生成的遮掩数据
	count := 3 * 1024 * 1024 / (bitLen / 8)

	fmt.Println("Count", count)
	k := make([]*big.Int, count)
	for i := 0; i < count; i++ {
		k[i] = new(big.Int)
	}

	k[count-1].SetBytes(stripeID).Mod(k[count-1], order)

	if k[count-1].Bit(0) == 0 {
		k[count-1].Or(k[count-1], bigOne)
	}

	k[count-1].Exp(k[count-1], v, order)

	tmp := new(big.Int)
	for l := 0; l < layers-1; l++ {
		fmt.Println("开始生成层:", l)
		if l%2 == 0 {
			k[0] = k[count-1]
			s1 := rand.NewSource(k[0].Int64())
			r1 := rand.New(s1)
			for i := 1; i < count; i++ {
				par := genParents(i, 2, k[0], r1)
				for _, j := range par {
					k[i].Add(k[i], k[j])
				}

				h1 := sha256.New()
				h1.Write(k[i].Bytes())
				tmp.SetBytes(h1.Sum(nil))

				if tmp.Bit(0) == 0 {
					tmp.Or(tmp, bigOne)
				}

				tmp.Exp(tmp, v, order)
				k[i].Add(k[i], tmp)
				k[i].Mod(k[i], order)
			}
		} else {
			k[count-1] = k[0]
			s1 := rand.NewSource(k[count-1].Int64())
			r1 := rand.New(s1)
			for i := 1; i < count; i++ {
				par := genParents(i, 2, k[0], r1)
				for _, j := range par {
					k[count-1-i].Add(k[count-1-i], k[count-1-j])
				}

				h1 := sha256.New()
				h1.Write(k[count-1-i].Bytes())
				tmp.SetBytes(h1.Sum(nil))

				if tmp.Bit(0) == 0 {
					tmp.Or(tmp, bigOne)
				}

				tmp.Exp(tmp, v, order)
				k[count-1-i].Add(k[count-1-i], tmp)
				k[count-1-i].Mod(k[count-1-i], order)
			}
		}

	}

	fmt.Println("last layer")
	for i := count - 2; i >= 0; i-- {
		k[i] = k[i].Add(k[i], k[i+1]).Mod(k[i], order).Xor(k[i], k[count-i-2])
	}

	return data
}

func encode(stripeID []byte, data []byte, order *big.Int, bitLen, dif int, isEncode bool) []byte {
	v := big.NewInt(1).Lsh(big.NewInt(1), uint(dif))
	bigOne := new(big.Int).SetInt64(1)
	//需要生成的遮掩数据
	count := 3 * 1024 * 1024 / (bitLen / 8)

	fmt.Println("Count", count)
	k := make([]*big.Int, count)
	k[0] = new(big.Int)
	stripeID = []byte("1234567890987654321")
	k[0].SetBytes(stripeID).Mod(k[0], order)

	if k[0].Bit(0) == 0 {
		k[0].Or(k[0], bigOne)
	}

	s1 := rand.NewSource(k[0].Int64())
	r1 := rand.New(s1)

	k[0].Exp(k[0], v, order)

	fmt.Println("开始生成掩盖数据")
	for i := 1; i < count; i++ {
		k[i] = big.NewInt(int64(i))

		par := genParents(i, 2, k[0], r1)
		for _, j := range par {
			k[i].Add(k[i], k[j])
		}

		if k[i].Bit(0) == 0 {
			k[i].Or(k[i], bigOne)
		}

		k[i].Mod(k[i], order)
		k[i].Exp(k[i], v, order)
	}
	fmt.Println("下一层")
	for i := count - 2; i >= 0; i-- {
		k[i] = k[i].Add(k[i], k[i+1]).Xor(k[i], k[count-i-2]).Mod(k[i], order)
	}
	home := os.Getenv("HOME")

	fmt.Println("开始异或数据")
	if isEncode {
		//生成第一个副本并写入
		f, err := os.Create(path.Join(home, "testFile-r1.data"))
		if err != nil {
			panic(err)
		}
		for i := 0; i < count/3; i++ {
			ans := xor(data[i*bitLen/8:(i+1)*bitLen/8], k[i].Bytes())
			f.Write(ans)
		}
		f.Close()

		f, err = os.Create(path.Join(home, "testFile-r2.data"))
		if err != nil {
			panic(err)
		}
		for i := count / 3; i < count/3*2; i++ {
			ans := xor(data[(i-count/3)*bitLen/8:(i+1-count/3)*bitLen/8], k[i].Bytes())
			f.Write(ans)
		}
		f.Close()

		f, err = os.Create(path.Join(home, "testFile-r3.data"))
		if err != nil {
			panic(err)
		}
		for i := count / 3 * 2; i < count; i++ {
			ans := xor(data[(i-count/3*2)*bitLen/8:(i+1-count/3*2)*bitLen/8], k[i].Bytes())
			f.Write(ans)
		}
		f.Close()
	} else {
		//生成第一个副本并写入
		f, err := os.Create(path.Join(home, "testFile-f1.data"))
		if err != nil {
			panic(err)
		}
		for i := 0; i < count/3; i++ {
			ans := xor(data[i*bitLen/8:(i+1)*bitLen/8], k[i].Bytes())
			f.Write(ans)
		}
		f.Close()
	}
	return data
}

func genParents(index int, paNums int, seed *big.Int, r1 *rand.Rand) []int {
	result := make([]int, paNums)
	if index == 0 {
		for i := 0; i < paNums; i++ {
			result[i] = 0
		}
		return result
	}

	logi := int(math.Floor(math.Log2(float64(index * paNums))))
	for i := 0; i < paNums; i++ {
		if i == 0 {
			result[i] = index - 1
			continue
		}
		j := r1.Intn(logi)
		jj := 1 << uint(j+1)
		if jj > index*paNums+i {
			jj = index*paNums + i
		}

		jj1 := jj >> 1
		if jj1 < 2 {
			jj1 = 2
		}

		divs := jj + 1 - jj1
		res := r1.Intn(divs)
		res += jj1
		if res >= index {
			result[i] = index - 1
		} else {
			result[i] = res
		}

	}
	return result
}

//注意，默认a的长度一定大于等于b
func xor(a, b []byte) []byte {
	data := make([]byte, len(a))
	diff := len(a) - len(b)
	for i := 0; i < diff; i++ {
		data[i] = a[i]
	}
	for i := diff; i < len(a); i++ {
		data[i] = a[i] ^ b[i-diff]
	}
	return data
}
