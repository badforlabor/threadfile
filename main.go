/**
 * Auth :   liubo
 * Date :   2019/12/2 18:24
 * Comment:
 */

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ChunkSize           = 1024 * 32
	worker = 100
	testCount = 10
)

func main() {
	//var src = "F:/workspace2/nomanscity-eng/CityLevel_Finish/Content/Movies/mp4.mp4"
	var src = "D:/workspace3/psl/PSL/Saved/StagedBuilds/WindowsNoEditor/PSLVR/Content/Paks/PSLVR-WindowsNoEditor.pak"
	var dst = "debug.pak"

	srcFile, _ := os.Open(src)
	if srcFile == nil {
		return
	}

	srcFileInfo, _ := os.Stat(src)
	if srcFileInfo == nil {
		return
	}

	for i:=0; i<testCount; i++ {
		b := testThreadFile(srcFile, srcFileInfo.Size(), dst)
		if b {
			fmt.Println("测试结果，成功", i)
		} else {
			fmt.Println("测试结果，失败", i)
			break
		}
	}
}

func testThreadFile(srcFile *os.File, fileSize int64, dst string) bool {

	atomic.StoreInt32(&workerDone, 0)

	os.Remove(dst)
	dstFile, _ := os.Create(dst)
	dstFile.Truncate(fileSize)
	dstFile.Sync()

	var chunkCount = (fileSize + ChunkSize - 1) / ChunkSize
	var workCount = chunkCount / worker
	for i:=int64(0); i<worker; i++ {
		go threadWriteFile(i*workCount, workCount, srcFile, dstFile)
	}
	go threadWriteFile(worker* workCount, chunkCount - worker* workCount, srcFile, dstFile)

	for {
		time.Sleep(time.Second)
		if workerDone == worker + 1 {
			break
		}
	}

	dstFile.Sync()

	return compareFile(srcFile, dstFile)
}
var fileMutex sync.Mutex
var workerDone int32
func threadWriteFile(idx int64, cnt int64, srcFile *os.File, dstFile *os.File) {
	buff := make([]byte, ChunkSize)

	for i:=int64(0); i<cnt; i++ {
		// io是线程暗转的
		//fileMutex.Lock()
		var offset = (idx + i) * ChunkSize
		n, err := srcFile.ReadAt(buff, offset)
		time.Sleep(time.Microsecond * time.Duration( rand.Int31n(10000)))
		if err == nil || err == io.EOF {
			dstFile.WriteAt(buff[0:n], offset)
		} else {
			fmt.Println("error", err.Error())
		}
		//fileMutex.Unlock()
	}

	atomic.AddInt32(&workerDone, 1)
}

func compareFile(f1, f2 *os.File) bool {
	count := ChunkSize// 32 * 1024
	buff1 := make([]byte, count)
	buff2 := make([]byte, count)

	f1.Seek(0, io.SeekStart)
	f2.Seek(0, io.SeekStart)
	{
		offset1, _ := f1.Seek(0, os.SEEK_CUR)
		offset2, _ := f2.Seek(0, os.SEEK_CUR)
		fmt.Println("比较文件：", offset1, offset2)
	}

	var equal = true
	for {
		n1, err1 := f1.Read(buff1)
		n2, err2 := f2.Read(buff2)
		if err1 != nil || err2 != nil {
			if err1 == io.EOF && err2 == io.EOF {

			} else {
				fmt.Println("1")
				equal = false
				break
			}
		}
		b := bytes.Equal(buff1[:n1], buff2[:n2])
		offset1, _ := f1.Seek(0, os.SEEK_CUR)
		offset2, _ := f2.Seek(0, os.SEEK_CUR)
		if !b {
			fmt.Println("2", offset1, offset2)
			b = true
			for i:=0; i<n1; i++ {
				if buff1[i] != buff2[i] {
					b = false
					break
				}
			}
			if !b {
				fmt.Println("二次比较失败！")
				ioutil.WriteFile("./test/" + strconv.Itoa(int(offset1)) + "_src", buff1[:n1], os.ModePerm)
				ioutil.WriteFile("./test/" + strconv.Itoa(int(offset2)) + "_dst", buff1[:n2], os.ModePerm)
				equal = false
			}
			continue
		}

		if n1 < count {
			if !equal {
				break
			}
			fmt.Println("两个文件相同！")
			return true
		}
	}
	return equal
}
