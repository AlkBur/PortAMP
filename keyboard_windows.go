package main

/*
#include <conio.h>
#include <stdio.h>
#include <stdlib.h>
int kbh(){
	return _kbhit();
}
int gch(){
	return _getch();
}
void clear(){
	fflush(stdin);
}
*/
import "C"
import (
	"time"
	"fmt"
)

const (
	KeyUP    int = 72 + 255
	KeyDOWN  int = 80 + 255
	KeyLEFT  int = 75 + 255
	KeyRIGHT int = 77 + 255
	KeyENTER int = 13
	KeySPACE int = 32
	KeyESC   int = 27
)

func GetKeyCode() int {
	time.Sleep(80 * time.Millisecond)
	C.clear()
	for {
		if C.kbh() != 1  {
			time.Sleep(80 * time.Millisecond)
			continue
		}
		rtn := int(C.gch())
		if rtn == 224 {
			rtn = int(C.gch()) + 255
		}
		C.clear()
		return rtn
	}
}

func WaitKeyCode(key chan int)  {
	for {
		i := GetKeyCode()
		key <- i
		fmt.Println(i)
	}
}
