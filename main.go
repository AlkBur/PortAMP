package main

import (
	"flag"
	"fmt"
	"os"
	"log"
	"os/signal"
	"syscall"
	"PortAMP/blob"
	"PortAMP/decoder"
	"time"
)

const PORTAMP_VERSION = "0.0.1";

var (
	debug bool
	enableTest bool
)

func init() {
	flag.StringVar(&cfg.m_OutputFile, `output-file`, "", `Config file`)
	flag.BoolVar(&cfg.m_Loop, `loop`, false, `Loop`)
	flag.BoolVar(&cfg.m_Verbose, `verbose`, false, `verbose`)
	flag.Var(&playlist.FileNames, "list", "List files.")
	flag.BoolVar(&debug, "debug", false,"Debug")
	flag.BoolVar(&enableTest, "test", false,"Run with test file")

	flag.Parse()
}

func main() {
	if len(os.Args) < 2 {
		PrintBanner()
		if !enableTest{
			return
		}
	}
	if err := cfg.ReadConfig(); err != nil{
		log.Fatal(err)
	}
	if enableTest {
		if playlist.IsEmpty() {
			playlist.EnqueueTrack( "birds.wav" )
			playlist.EnqueueTrack( "D:/Музыка/100 Krasiveishih Pesen 2015/01. Gary Moore - Still Got The Blues.mp3" )
		}
		cfg.m_Verbose = true
	}

	audio, err := NewAudioSystem()
	if err != nil {
		log.Fatal(err)
	}
	defer audio.Close()
	audio.Start()

	Source := audio.CreateAudioSource()
	defer Source.Close()
	if ( playlist.GetNumTracks() == 1 ) {
		Source.SetLooping( cfg.m_Loop )
	}


	key_press_chan := make(chan int, 1)
	exit_chan := make(chan struct{}, 1)

	//Запускаем обработчики сигналов системы
	go WaitKeyCode(key_press_chan)
	handleSignal(exit_chan)

	//Запускаем прогрыватель
	RequestingExit := false
	for !playlist.IsEmpty() && !RequestingExit {
		file := playlist.GetAndPopNextTrack(cfg.m_Loop)
		DataBlob, err := blob.ReadFile(file)
		if err != nil || DataBlob.GetDataSize()==0 {
			log.Println(err)
			continue
		}
		Provider, err := decoder.New(DataBlob, IsVerbose())
		if err != nil {
			log.Println(err)
			continue
		}
		if err = Source.BindDataProvider( Provider ); err != nil {
			log.Println(err)
			continue
		}

		bContinue := true

		err = Source.Play()
		if err != nil {
			log.Println(err)
			bContinue = false
		}

		//ожидаем сигнаов от системы
		for bContinue {
			select {
			case <-key_press_chan:
				bContinue = false
				RequestingExit = true
			case <-exit_chan:
				RequestingExit = true
				bContinue = false
			case <-time.After(time.Millisecond * 100):
				if !Source.IsPlaying(){
					log.Println("is not play")
					bContinue = false
				}
			}
		}
		log.Println("Stop play")
		Source.Stop()
	}
}

func handleSignal(stop_play chan struct{}) {
	go func() {
		c := make(chan os.Signal)

		signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		s := <-c
		log.Printf("Got signal [%s]\n", s)
		stop_play <- struct{}{}
	}()
}

func PrintBanner() {
	fmt.Printf( "PortAMP version %s\n", PORTAMP_VERSION)
	fmt.Printf( "Copyright (C) 2018-2018 Burlakov Alexander\n" );
	fmt.Printf( "\n" );
	fmt.Printf( "portamp <filename1> [<filename2> ...] [--loop] [--wav-modplug] [--verbose]\n" );
	fmt.Printf( "\n" );
}

