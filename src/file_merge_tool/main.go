package main

import (
	"fmt"
	"flag"
	"os"
	"strconv"
)

const (
	fileNr = 2
	defPadding  = 0xFF
	version = "V1.0.0"
)

var (
	desFile string
	padding int
	debug bool
	board string
)

type fileDesc struct {
	path string
	buf []byte
	splitSize int
	err error
}

var fileTable [fileNr]fileDesc

func main() {
	fmt.Println("=======================================")
	fmt.Println("=>  binary file merge tool.", version)
	fmt.Println("=>  author: hgqian")
	fmt.Println("=>  e-mail: hgq819@126.com")
	fmt.Println("=======================================")

	for i,_ := range fileTable {
		num := strconv.Itoa(i+1)
		flag.StringVar(&fileTable[i].path, "f" + num, "", "merge source file " + num + ".")

		switch i {
		case 0: flag.IntVar(&fileTable[i].splitSize, "s" + num, 0x4000, "split the size of file "+num+". default(16k)")
		case 1: flag.IntVar(&fileTable[i].splitSize, "s" + num, 0xc000, "split the size of file "+num+". default(48k)")
		default:
			flag.IntVar(&fileTable[i].splitSize, "s"+num, 0, "split the size of file "+num+".")
		}
	}
	flag.StringVar(&desFile, "target", "merge_file.bin", "target file path.")
	flag.StringVar(&desFile, "t", "merge_file.bin", "< --target >")
	flag.IntVar(&padding, "p", defPadding, "padding char default: <255:0xFF>.")
	flag.BoolVar(&debug, "d", false, "print debug info.")
	flag.StringVar(&board, "b", "051", "051 or 071. preset split size value of s1 s2.")

	flag.Parse()

	switch board {
	case "051":
		fileTable[0].splitSize = 0x4000
		fileTable[1].splitSize = 0xC000
	case "071":
		fileTable[0].splitSize = 0x4000
		fileTable[1].splitSize = 0x1C000
	}

	flag := false
	for i, _ := range fileTable {
		fmt.Printf("name:%s    size:0x%x\r\n", fileTable[i].path, fileTable[i].splitSize)
		if (fileTable[i].path != "") && (fileTable[i].splitSize > 0) {

			if isExit,err := pathExists(fileTable[i].path); isExit == false {
				fmt.Println(err)
				return
			}

			fileTable[i].buf, fileTable[i].err = loadFile(fileTable[i].path)
			{
				if fileTable[i].err != nil {
					fmt.Println(fileTable[i].err)
					return
				}
			}
			fileTable[i].buf = resizeBuf(fileTable[i].buf, fileTable[i].splitSize, byte(padding))
			flag = true
		}
		if debug {
			/*
			for _,v := range fileTable[i].buf {
				fmt.Printf("%02x ", v)
			}
			*/
			fd, err := os.Create("temp" + strconv.Itoa(i+1) + ".bin")
			if err != nil {
				fmt.Println(err)
				return
			}
			defer fd.Close()
			for i,_ := range fileTable {
				l, err := fd.Write(fileTable[i].buf)
				if err != nil {
					fmt.Println(err)
					return
				}
				if l != len(fileTable[i].buf) {
					fmt.Printf("error: write fail. exp:%d real:%d\n", len(fileTable[i].buf), l)
					return
				}
			}
		}
		fmt.Println("---------------------------")
	}

	if flag == false {
		return
	}

	fd, err := os.Create(desFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fd.Close()
	for i,_ := range fileTable {
		l, err := fd.Write(fileTable[i].buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		if l != len(fileTable[i].buf) {
			fmt.Printf("error: write fail. exp:%d real:%d\n", len(fileTable[i].buf), l)
			return
		}
	}
}

/*
 *	加载文件
 *
 *	buf -> 读取的文件内容
 *  err -> 错误信息
 */
func loadFile(filePath string)(buf []byte, err error) {
	var fd *os.File
	var info os.FileInfo
	var size int
	err = nil

	if filePath != "" {
		defer fd.Close()
		fd, err = os.OpenFile(filePath, os.O_RDONLY, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		info, err = fd.Stat()
		if err != nil {
			fmt.Println(err)
		}

		//fmt.Println(info)

		buf = make([]byte, info.Size());
		size, err = fd.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("read size: %d byte 0x%x\n", size, size)
	}
	return
}

/*
 *  重新调整buffer，使其大小符合需求
 *
 *	buf 	-> 原始slice
 *	size 	-> 需要截取或者填充到的大小
 *  padChar -> 如果长度不足，用于填充的字符
 *
 *  return:
 *  []byte  -> 返回调整后的slice
 */
func resizeBuf(buf []byte, size int, padChar byte)([]byte){
	switch {
	case len(buf) < size:
		fmt.Println("warning: Padding the file.")
		pad := make([]byte, size - len(buf))
		for i,_ := range pad {
			pad[i] = byte(padChar)
		}
		buf = append(buf, pad...)

	case len(buf) > size:
		fmt.Println("warning: Split the file.")
		buf = buf[0:size]
	default:

	}
	return buf
}


func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	/*
	if os.IsNotExist(err) {
		return false, err
	}
	*/
	return false, err
}