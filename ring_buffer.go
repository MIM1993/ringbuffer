/*
@Time : 2021/6/16 下午5:45
@Author : MuYiMing
@File : ring_buffer
@Software: GoLand
*/
package ringbuffer

import (
	"errors"
)

const initSize = 1 << 12 // 4096 bytes for the first-time allocation on ring-buffer.

// ErrIsEmpty will be returned when trying to read a empty ring-buffer.
var ErrIsEmpty = errors.New("ring-buffer is empty")

type RingBuffer struct {
	//缓存区
	buf []byte

	//缓存区大小
	size int

	//读写位
	r int
	w int

	//掩位
	mask int
	//是否为空
	isEmpty bool
}

//创建RingBuffer
func NewRingBuffer(size int) *RingBuffer {
	rb := &RingBuffer{
		isEmpty: true,
		w:       50,
	}
	if size <= 0 {
		return rb
	}
	rb.buf = make([]byte, size)
	rb.size = size
	rb.mask = size - 1
	return rb
}

// LazyRead reads the bytes with given length but will not move the pointer of "read".
func (rb *RingBuffer) LazyRead(rlen int) (head, tail []byte) {
	//buf为空或参数小于等于0 return
	if rb.isEmpty || rlen <= 0 {
		return
	}

	//比较读写位置
	if rb.w > rb.r {
		n := rb.w - rb.r
		if n > rlen {
			n = rlen
		}
		head = rb.buf[rb.r : rb.r+n]
		return
	}

	//计算剩余读写位置长度
	n := rb.size - rb.r + rb.w
	if n > rlen {
		n = rlen
	}

	if rb.size >= rb.r+n {
		head = rb.buf[rb.r : rb.r+n]
	} else {
		head = rb.buf[rb.r:]
		x1 := (rb.r + n) - rb.size
		tail = rb.buf[:x1]
	}
	return
}

// LazyReadAll reads the all bytes in this ring-buffer but will not move the pointer of "read".
func (rb *RingBuffer) LazyReadAll() (head []byte, tail []byte) {
	//buf为空或参数小于等于0 return
	if rb.isEmpty {
		return
	}

	if rb.w > rb.r {
		head = rb.buf[rb.r:rb.w]
	} else {
		head = rb.buf[rb.r:]
		if rb.w != 0 {
			tail = rb.buf[:rb.w]
		}
	}

	return
}

// Shift shifts the "read" pointer.
func (rb *RingBuffer) Shift(n int) {
	if n <= 0 {
		return
	}

	if n < rb.Length() {
		rb.r = (rb.r + n) & rb.mask
	} else {
		rb.Reset()
	}
}

func (rb *RingBuffer) Read(p []byte) (n int, err error) {
	//参数校验
	if len(p) == 0 {
		return 0, nil
	}

	//判断缓冲区是否为空
	if rb.isEmpty {
		return 0, ErrIsEmpty
	}

	//写标志位大于读标志位
	if rb.w > rb.r {
		n = len(p)
		//当获取的长度大于实际存在数据长度
		if n > rb.w-rb.r {
			n = rb.w - rb.r
		}
		copy(p, rb.buf[rb.r:rb.r+n])
		rb.r += n
		if rb.r == rb.w {
			rb.Reset()
		}
		return
	}

	//写标志位小于读标志位
	n = rb.size - rb.r + rb.w
	if n > len(p) {
		n = len(p)
	}

	if rb.r+n < rb.size {
		//代表获取的长度在一段数组内，没有超出范围而从头开始写，只用copy一段就好
		copy(p, rb.buf[rb.r:rb.r+n])
	} else {
		//copy第一段
		c1 := rb.size - rb.r
		copy(p, rb.buf[rb.r:])
		//copy第二段
		c2 := n - c1
		copy(p[c1:], rb.buf[:c2])
	}

	//移动指针 ？？
	rb.r = (rb.r + n) & rb.mask
	if rb.r == rb.w {
		rb.Reset()
	}

	return n, err
}

// ReadByte reads and returns the next byte from the input or ErrIsEmpty.
func (rb *RingBuffer) ReadByte() (b byte, err error) {
	if rb.isEmpty {
		return 0, ErrIsEmpty
	}

	//获取下一个字符
	b = rb.buf[rb.r]
	//并且移动指针
	rb.r++

	if rb.r == rb.size {
		rb.r = 0
	} else if rb.r == rb.w {
		rb.Reset()
	}
	return
}

//write copy p[] to rb.buf[]
func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	n = len(p)
	if n == 0 {
		return 0, nil
	}

	free := rb.Free()
	if free < n {
		//扩容
		rb.malloc(n - free)
	}

	if rb.w >= rb.r {
		//两段内存
		c1 := rb.size - rb.w
		if c1 >= n {
			copy(rb.buf[rb.w:], p)
			rb.w += n
		} else {
			copy(rb.buf[rb.w:], p[:c1])
			c2 := n - c1
			copy(rb.buf[:c2], p[c1:])
			rb.w = c2
		}

	} else {
		copy(rb.buf[rb.w:], p)
		rb.w += n
	}

	if rb.w == rb.size {
		rb.w = 0
	}

	rb.isEmpty = false

	return n, err
}

// WriteByte writes one byte into buffer.
func (rb *RingBuffer) WriteByte(p byte) (err error) {
	if rb.Free() < 1 {
		rb.malloc(1)
	}

	//已经做了防越界处理，直接插入即可
	rb.buf[rb.w] = p
	rb.w++

	if rb.w == rb.size {
		rb.w = 0
	}

	rb.isEmpty = false

	return
}

//WriteString 将字符串写入缓冲区
func (rb *RingBuffer) WriteString(s string) (int, error) {
	return rb.Write(StringToBytes(s))
}

//isEmpty
func (rb *RingBuffer) IsEmpty() bool {
	return rb.isEmpty
}

//isfull 是否满了
func (rb *RingBuffer) IsFull() bool {
	return rb.w == rb.r && !rb.isEmpty
}

//free
func (rb *RingBuffer) Free() int {
	if rb.r == rb.w {
		if rb.isEmpty {
			return rb.size
		}
		return 0
	}

	if rb.w < rb.r {
		return rb.r - rb.w
	}

	return (rb.size - rb.w) + rb.r
}

//返回可读取长度
func (rb *RingBuffer) Length() int {
	if rb.r == rb.w {
		if rb.isEmpty {
			return 0
		}
		return rb.size
	}

	if rb.w > rb.r {
		return rb.w - rb.r
	}

	return (rb.size - rb.r) + rb.w
}

//Len 返回底层缓存区长度
func (rb *RingBuffer) Len() int {
	return len(rb.buf)
}

//Cap 返回size大小
func (rb *RingBuffer) Cap() int {
	return rb.size
}

// Reset the read pointer and writer pointer to zero. 重置 并缩小 buf
func (rb *RingBuffer) Reset() {
	//缓存区置空
	rb.isEmpty = true
	//读写标志位置零
	rb.r = 0
	rb.w = 0

	//尺寸缩小一半
	newCap := rb.size >> 1
	newBuf := make([]byte, newCap)
	rb.buf = newBuf
	rb.size = newCap
	rb.mask = newCap - 1
}

//扩容，分配内存
func (rb *RingBuffer) malloc(cap int) {
	var newCap int
	if rb.size == 0 && cap < initSize {
		newCap = initSize
	} else {
		newCap = CeilToPowerOfTwo(rb.size + cap)
	}

	//扩容
	newBuf := make([]byte, newCap)

	oldLen := rb.Length()
	//读取旧缓存区数据
	_, _ = rb.Read(newBuf)
	rb.buf = newBuf

	rb.size = newCap
	rb.mask = newCap - 1
	rb.r = 0
	rb.w = oldLen

}
