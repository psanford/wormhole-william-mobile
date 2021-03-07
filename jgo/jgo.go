package jgo

/*
#include <jni.h>
*/
import "C"

import (
	"errors"
	"log"
	"sync"
	"unsafe"

	"gioui.org/app"
	"git.wow.st/gmp/jni"
)

var (
	pendingResultMux sync.Mutex
	pendingResult    chan PickResult
)

type PickResult struct {
	Path string
	Name string
	Err  error
}

func PickFile(viewEvt app.ViewEvent) <-chan PickResult {
	pendingResultMux.Lock()
	if pendingResult != nil {
		pendingResult <- PickResult{
			Err: errors.New("New PickFile has taken precidence"),
		}
	}
	pendingResult = make(chan PickResult, 1)
	pendingResultMux.Unlock()

	go func() {
		jvm := jni.JVMFor(app.JavaVM())
		err := jni.Do(jvm, func(env jni.Env) error {

			var uptr = app.AppContext()
			appCtx := *(*jni.Object)(unsafe.Pointer(&uptr))
			loader := jni.ClassLoaderFor(env, appCtx)
			cls, err := jni.LoadClass(env, loader, "io.sanford.wormholewilliam.Jni")
			if err != nil {
				log.Printf("Load io.sanford.wormholewilliam.Jni error: %s", err)
			}

			mid := jni.GetMethodID(env, cls, "<init>", "()V")

			inst, err := jni.NewObject(env, cls, mid)
			if err != nil {
				log.Printf("NewObject err: %s", err)
			}

			mid = jni.GetMethodID(env, cls, "register", "(Landroid/view/View;)V")

			jni.CallVoidMethod(env, inst, mid, jni.Value(viewEvt.View))
			return err
		})

		if err != nil {
			log.Printf("Err: %s", err)
		}
	}()

	return pendingResult
}

//export Java_io_sanford_wormholewilliam_Jni_pickerResult
func Java_io_sanford_wormholewilliam_Jni_pickerResult(env *C.JNIEnv, cls C.jclass, jpath, jname, jerr C.jstring) {

	jenv := jni.EnvFor(uintptr(unsafe.Pointer(env)))

	path := jni.GoString(jenv, jni.String(jpath))
	name := jni.GoString(jenv, jni.String(jname))
	errStr := jni.GoString(jenv, jni.String(jerr))
	log.Printf("pickResult path: %s err: %s", path)

	result := PickResult{
		Path: path,
		Name: name,
	}
	if errStr != "" {
		result.Err = errors.New(errStr)
	}

	pendingResultMux.Lock()
	pendingResult <- result
	pendingResult = nil
	pendingResultMux.Unlock()
}
