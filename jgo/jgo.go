// +build android

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
	"github.com/psanford/wormhole-william-mobile/internal/picker"
)

var (
	pendingResultMux sync.Mutex
	pendingResult    chan picker.PickResult
)

func PickFile(viewEvt app.ViewEvent) <-chan picker.PickResult {
	pendingResultMux.Lock()
	if pendingResult != nil {
		pendingResult <- picker.PickResult{
			Err: errors.New("New PickFile has taken precidence"),
		}
	}
	pendingResult = make(chan picker.PickResult, 1)
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

			return jni.CallVoidMethod(env, inst, mid, jni.Value(viewEvt.View))
		})

		if err != nil {
			log.Printf("Err: %s", err)
		}
	}()
	return pendingResult
}

func NotifyDownloadManager(viewEvt app.ViewEvent, name, path, contentType string, size int64) {
	go func() {
		jvm := jni.JVMFor(app.JavaVM())
		err := jni.Do(jvm, func(env jni.Env) error {
			var uptr = app.AppContext()
			appCtx := *(*jni.Object)(unsafe.Pointer(&uptr))
			loader := jni.ClassLoaderFor(env, appCtx)
			cls, err := jni.LoadClass(env, loader, "io.sanford.wormholewilliam.Download")
			if err != nil {
				log.Printf("Load io.sanford.wormholewilliam.Download error: %s", err)
			}

			mid := jni.GetMethodID(env, cls, "<init>", "()V")

			inst, err := jni.NewObject(env, cls, mid)
			if err != nil {
				log.Printf("NewObject err: %s", err)
			}
			sig := "(Landroid/view/View;Ljava/lang/String;Ljava/lang/String;Ljava/lang/String;J)V"

			mid = jni.GetMethodID(env, cls, "register", sig)

			jname := jni.JavaString(env, name)
			jpath := jni.JavaString(env, path)
			jcontentType := jni.JavaString(env, contentType)

			return jni.CallVoidMethod(env, inst, mid, jni.Value(viewEvt.View), jni.Value(jname), jni.Value(jpath), jni.Value(jcontentType), jni.Value(size))
		})
		if err != nil {
			log.Printf("Err: %s", err)
		}
	}()
}

//export Java_io_sanford_wormholewilliam_Jni_pickerResult
func Java_io_sanford_wormholewilliam_Jni_pickerResult(env *C.JNIEnv, cls C.jclass, jpath, jname, jerr C.jstring) {

	jenv := jni.EnvFor(uintptr(unsafe.Pointer(env)))

	path := jni.GoString(jenv, jni.String(jpath))
	name := jni.GoString(jenv, jni.String(jname))
	errStr := jni.GoString(jenv, jni.String(jerr))
	log.Printf("pickResult path: %s err: %s", path)

	result := picker.PickResult{
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
