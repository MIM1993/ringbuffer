/*
@Time : 2021/6/16 下午5:46
@Author : MuYiMing
@File : help_test
@Software: GoLand
*/
package ringbuffer

import (
	"fmt"
	"testing"
)

func TestHelper(t *testing.T) {
	var tmp int = 12423452
	tmp1 := fillBits(tmp)
	fmt.Println(tmp1)
}

func TestOther(t *testing.T) {
	fmt.Println(1 << 12)

	fmt.Println(bitsize)
	fmt.Println(maxintHeadBit)
	fmt.Println(^uint(0) >> 63)
	fmt.Println(32 << 0)
	fmt.Println(32 << (^uint(0) >> 63))

}

func TestOther1(t *testing.T) {
	//var a,b int = 120,99
	//
	//c := a & b

	//fmt.Printf("c = %d\n",c)

	//rb := NewRingBuffer(100)
	//fmt.Printf("%+v\n",rb)
	//rb.Shift(10)
	//fmt.Printf("%+v\n",rb)

	//fmt.Println(maxintHeadBit)
	//n := 100345634560
	//fmt.Println(n & maxintHeadBit)

	fmt.Println(CeilToPowerOfTwo(99))
	fmt.Println(CeilToPowerOfTwo(129))


}
