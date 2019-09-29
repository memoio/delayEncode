package delayencode

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

func BenchmarkEncode(b *testing.B) {
	rand.Seed(time.Now().Unix())
	data := make([]byte, dataSize)
	fmt.Println("分配内存完毕")
	fillRandom(data)
	home := os.Getenv("HOME")
	f, _ := os.Create(path.Join(home, "testFile-raw.data"))
	f.Write(data)
	fmt.Println("数据生成完毕")
	order, _, _ := GenParams(m)
	stripeID := make([]byte, 32)
	fillRandom(stripeID)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encode(stripeID, data, order, m, difficulty, true)
	}
}

func BenchmarkLayerEncode(b *testing.B) {
	//在此调整一些参数
	dataSize = 1024 * 1024
	m = 128 * 8 * 4
	difficulty = 4
	layers = 11
	conn = 2
	//在此打印一些参数
	fmt.Println("编码输出大小:", 3*dataSize/(1024*1024), "M")
	fmt.Println("GF(2^m) m =", m, ",", m/8, "byte")
	fmt.Println("layers =", layers)
	fmt.Println("难度值", difficulty)
	fmt.Println("DRG依赖节点数 = ", conn)

	rand.Seed(time.Now().Unix())
	data := make([]byte, dataSize)
	fmt.Println("分配内存完毕")
	fillRandom(data)
	home := os.Getenv("HOME")
	f, _ := os.Create(path.Join(home, "testFile-raw.data"))
	f.Write(data)
	fmt.Println("数据生成完毕")
	order, _, _ := GenParams(m)
	stripeID := make([]byte, 32)
	fillRandom(stripeID)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layerEncode(stripeID, data, order, m, difficulty, layers)
	}
}

func BenchmarkGenParaent(b *testing.B) {
	k := big.NewInt(int64(123456789))
	s1 := rand.NewSource(k.Int64())
	r1 := rand.New(s1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k := big.NewInt(int64(i))
		genParents(i, 8, k, r1)
	}
}

func BenchmarkDifficulty(b *testing.B) {
	dif := 4
	v := big.NewInt(1).Lsh(big.NewInt(1), uint(dif))

	h1 := sha256.New()
	h1.Write([]byte("123456"))
	k := new(big.Int)
	k.SetBytes(h1.Sum(nil))
	order, _, _ := GenParams(32 * 8 * 16)
	bigOne := new(big.Int).SetInt64(1)
	if k.Bit(0) == 0 {
		k.Or(k, bigOne)
	}
	k.Exp(k, v, order)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := big.NewInt(int64(i))
		tmp.Add(tmp, k)
		if tmp.Bit(0) == 0 {
			tmp.Or(tmp, bigOne)
		}
		tmp.Exp(tmp, v, order)
	}
}

func Benchmark32BWithHash(b *testing.B) {
	dif := 4
	v := big.NewInt(1).Lsh(big.NewInt(1), uint(dif))

	h1 := sha256.New()
	h1.Write([]byte("123456"))
	k := new(big.Int)
	k.SetBytes(h1.Sum(nil))
	order, _, _ := GenParams(32 * 8 * 16)

	bigOne := new(big.Int).SetInt64(1)
	if k.Bit(0) == 0 {
		k.Or(k, bigOne)
	}

	k.Exp(k, v, order)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := big.NewInt(int64(i))
		tb := tmp.Add(tmp, k).Bytes()
		h1 := sha256.New()
		h1.Write(tb)
		tmp.SetBytes(h1.Sum(nil))
		if tmp.Bit(0) == 0 {
			tmp.Or(tmp, bigOne)
		}
		tmp.Exp(tmp, v, order)
		tmp.Add(tmp, k).Mod(tmp, order)
	}
}

func BenchmarkHash(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h1 := sha256.New()
		ib := strconv.Itoa(i)
		h1.Write([]byte(ib))
		h1.Sum(nil)
		//tmp := new(big.Int)
		//tmp.SetBytes(res)
		//tmp.Bytes()
	}
}

func Benchmark32BExp(b *testing.B) {
	dif := 4
	v := big.NewInt(1).Lsh(big.NewInt(1), uint(dif))

	h1 := sha256.New()
	h1.Write([]byte("123456"))
	k := new(big.Int)
	k.SetBytes(h1.Sum(nil))
	order, _, _ := GenParams(32 * 8 * 16)
	bigOne := new(big.Int).SetInt64(1)
	if k.Bit(0) == 0 {
		k.Or(k, bigOne)
	}
	//k.Exp(k, v, order)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := big.NewInt(int64(i))
		tmp.Add(tmp, k)
		if tmp.Bit(0) == 0 {
			tmp.Or(tmp, bigOne)
		}
		tmp.Exp(tmp, v, order)
	}
}

func BenchmarkAdd(b *testing.B) {
	dif := 4
	v := big.NewInt(1).Lsh(big.NewInt(1), uint(dif))

	h1 := sha256.New()
	h1.Write([]byte("123456"))
	k := new(big.Int)
	k.SetBytes(h1.Sum(nil))
	order, _, _ := GenParams(32 * 8 * 16)
	k.Exp(k, v, order)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := big.NewInt(int64(i))
		tmp.Add(tmp, k).Mod(tmp, order)
	}
}

func BenchmarkXor(b *testing.B) {
	dif := 4
	v := big.NewInt(1).Lsh(big.NewInt(1), uint(dif))

	h1 := sha256.New()
	h1.Write([]byte("123456"))
	k := new(big.Int)
	k.SetBytes(h1.Sum(nil))
	order, _, _ := GenParams(32 * 8 * 16)
	k.Exp(k, v, order)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := big.NewInt(int64(i))
		tmp.Add(tmp, k).Mod(tmp, order)
		tmp.Xor(tmp, k)
	}
}

func fillRandom(p []byte) {
	for i := 0; i < len(p); i += 7 {
		val := rand.Int63()
		for j := 0; i+j < len(p) && j < 7; j++ {
			p[i+j] = byte(val)
			val >>= 8
		}
	}
}

func BenchmarkNewBigInt(b *testing.B) {
	count := 3

	fmt.Println("Count", count)
	k := make([]*big.Int, count)
	for i := 0; i < count; i++ {
		k[i] = big.NewInt(int64(i))
		fmt.Println(k[i])
	}

	tmp := big.NewInt(5)
	k[0] = k[1]

	k[1] = tmp

	for i := 0; i < count; i++ {
		fmt.Println(k[i])
	}
}
