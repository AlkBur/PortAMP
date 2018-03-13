package main

/*
#include <termios.h>
#include <unistd.h>

/*
#include <termios.h>
#include <unistd.h>
#include <stdio.h>
int getch() {
	int ch;
    struct termios t_old, t_new;
    tcgetattr(STDIN_FILENO, &t_old);
    t_new = t_old;
    t_new.c_lflag &= ~(ICANON | ECHO);
    tcsetattr(STDIN_FILENO, TCSANOW, &t_new);
    ch = getchar();
    tcsetattr(STDIN_FILENO, TCSANOW, &t_old);
    return ch;
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
	rtn := int(C.getch())
	return rtn
}

func WaitKeyCode(key chan int)  {
	for {
		i := GetKeyCode()
		key <- i
		fmt.Println(i)
	}
}
