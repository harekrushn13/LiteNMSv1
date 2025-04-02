package profiling

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
)

func StartCPUProfile() *os.File {

	f, err := os.Create("./profiling/profiles/cpu.prof")

	if err != nil {

		log.Fatal("could not create CPU profile: ", err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {

		log.Fatal("could not start CPU profile: ", err)
	}

	return f
}

func StopCPUProfile(f *os.File) {
	pprof.StopCPUProfile()

	f.Close()
}

func WriteMemProfile() {

	f, err := os.Create("./profiling/profiles/mem.prof")

	if err != nil {

		log.Fatal("could not create memory profile: ", err)
	}

	defer f.Close()

	runtime.GC()

	if err := pprof.WriteHeapProfile(f); err != nil {

		log.Fatal("could not write memory profile: ", err)
	}
}

func WriteGoroutineProfile() {

	f, err := os.Create("./profiling/profiles/goroutine.prof")

	if err != nil {

		log.Fatal("could not create goroutine profile: ", err)
	}

	if err := pprof.Lookup("goroutine").WriteTo(f, 0); err != nil {

		log.Fatal("could not write goroutine profile: ", err)
	}

	f.Close()
}

func StartTrace() *os.File {

	f, err := os.Create("./profiling/profiles/trace.out")

	if err != nil {

		log.Fatal("could not create trace output file: ", err)
	}

	if err := trace.Start(f); err != nil {

		log.Fatal("could not start trace: ", err)
	}

	return f
}

func StopTrace(f *os.File) {

	trace.Stop()

	f.Close()
}
